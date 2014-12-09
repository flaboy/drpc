package drpc

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	apiHandlers map[string]ApiHandler
	OnConnect   func(*Connection)
	OnClose     func(string)
	Connections map[string]*Connection
}

func NewServer() *Server {
	s := &Server{
		apiHandlers: make(map[string]ApiHandler),
		Connections: make(map[string]*Connection),
	}
	s.apiHandlers["@set_my_id"] = s.setClientId
	return s
}

func (me *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ws, err := websocket.Upgrade(w, req, nil, 1024, 1024)
	if err != nil {
		log.Println("websocket-error:", err)
		return
	}

	me.wshandler(req, ws)
}

func (me *Server) wshandler(req *http.Request, ws *websocket.Conn) {
	ep := newConnection()
	ep.apiHandlers = me.apiHandlers
	ep.conn = ws

	id := ep.Id()
	me.Connections[id] = ep
	defer me.onClose(id)

	if me.OnConnect != nil {
		go me.OnConnect(ep)
	}

	ep.workloop()
}

func (me *Server) onClose(id string) {
	delete(me.Connections, id)
	if me.OnClose != nil {
		me.OnClose(id)
	}
}

func (me *Server) Handle(cmd string, f ApiHandler) {
	me.apiHandlers[cmd] = f
}

func (me *Server) setClientId(r *Request) Response {
	err := r.UnmarshalArgs(&r.Connection.client_id)
	return Response{Data: true, Err: err}
}
