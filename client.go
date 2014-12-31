package drpc

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	ep        *Connection
	Id        string
	OnConnect func()
	OnClose   func()
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

func (me *Client) Run(addr string) (err error) {
	for {
		me.run(addr)
		time.Sleep(time.Second)
	}
	return
}

func (me *Client) run(addr string) (err error) {
	var (
		use_ssl bool
		u       *url.URL
		c       net.Conn
		rsp     *http.Response
		conn    *websocket.Conn
	)

	u, err = url.Parse(addr)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return
	}

	if use_ssl {
		c = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	}

	headers := http.Header{}
	conn, rsp, err = websocket.NewClient(c, u, headers, 1024, 1024)
	if err != nil {
		log.Println(err, rsp)
		return
	}

	me.ep.conn = conn

	defer func() {
		if me.OnClose != nil {
			me.OnClose()
		}
	}()

	if me.OnConnect != nil {
		go me.OnConnect()
	}

	me.ep.workloop()
	return
}

func (me *Client) Handle(cmd string, f ApiHandler) {
	me.ep.handle(cmd, f)
}
