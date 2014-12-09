package drpc

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
)

type Client struct {
	ep *Connection
	Id string
}

func NewClient(id string) (c *Client) {
	c = &Client{Id: id}
	c.ep = newConnection()
	return c
}

func (me *Client) Channel() *Channel {
	if me.ep == nil {
		return nil
	}
	return me.ep.Channel()
}

func (me *Client) Connect(addr string) (err error) {
	var (
		use_ssl bool
		u       *url.URL
		c       net.Conn
		rsp     *http.Response
		conn    *websocket.Conn
	)

	u, err = url.Parse(addr)
	if err != nil {
		return
	}

	switch u.Scheme {
	case "http":
	case "ws":
		use_ssl = false
	case "https":
	case "wss":
		use_ssl = true
	}

	c, err = net.Dial("tcp", u.Host)
	if err != nil {
		return
	}

	if use_ssl {
		c = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	}

	headers := http.Header{}
	conn, rsp, err = websocket.NewClient(c, u, headers, 1024, 1024)
	if err != nil {
		fmt.Println(rsp)
		return
	}

	me.ep.conn = conn
	go me.ep.workloop()

	me.ep.Channel().Call("@set_my_id", me.Id)

	fmt.Sprintf("todo: ", use_ssl)

	return
}

func (me *Client) Handle(cmd string, f ApiHandler) {
	me.ep.handle(cmd, f)
}
