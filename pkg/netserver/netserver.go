package netserver

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"nhooyr.io/websocket"
)

type WebsocketTransport struct {
	onOpen  func(*Conn)
	onClose func(*Conn)
	onData  func(c *Conn, b []byte)
}

func (wt *WebsocketTransport) OnOpen(h func(c *Conn)) {
	wt.onOpen = h
}

func (wt *WebsocketTransport) OnClose(h func(c *Conn)) {
	wt.onClose = h
}

func (wt *WebsocketTransport) OnMessage(h func(c *Conn, b []byte)) {
	wt.onData = h
}

func (wt *WebsocketTransport) Connect(addr string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c, r, err := websocket.Dial(ctx, addr, nil)
	if err != nil {
		return err
	}

	conn := &Conn{
		addr: ReadUserIP(r.Request),
		c:    c,
	}

	wt.handleConn(conn)

	return nil
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	if IPAddress == "" {
		IPAddress = r.Host
	}
	return IPAddress
}

func (wt *WebsocketTransport) Handle(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// TODO: use proper origin whitelist
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}

	conn := &Conn{
		addr: ReadUserIP(r),
		c:    c,
	}

	go wt.handleConn(conn)
}

func (wt *WebsocketTransport) handleConn(conn *Conn) {
	if wt.onOpen != nil {
		wt.onOpen(conn)
	}

	for {
		mt, b, err := conn.c.Read(context.Background())
		if err != nil {
			if wt.onClose != nil {
				wt.onClose(conn)
			}

			return
		}

		if mt != websocket.MessageText {
			continue
		}

		if wt.onData != nil {
			wt.onData(conn, b)
		}
	}
}

type Conn struct {
	addr    string
	c       *websocket.Conn
	latency atomic.Int64
}

func (c *Conn) Addr() string {
	return c.addr
}

func (c *Conn) Ping() error {
	return c.c.Ping(context.Background())
}

func (c *Conn) Read() ([]byte, error) {
	mt, b, err := c.c.Read(context.Background())
	if err != nil {
		return nil, err
	}

	if mt != websocket.MessageText {
		return nil, errors.New("invalid message format")
	}

	return b, nil
}

func (c *Conn) Write(b []byte) error {
	return c.c.Write(context.Background(), websocket.MessageText, b)
}

func NewWebsocketTransport() *WebsocketTransport {
	return &WebsocketTransport{}
}
