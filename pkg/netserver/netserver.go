package netserver

import (
	"context"
	"log"
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
	c, r, err := websocket.Dial(context.Background(), addr, nil)
	if err != nil {
		log.Println("dial:", err)
		return err
	}

	conn := &Conn{
		addr: r.Request.RemoteAddr,
		c:    c,
	}

	wt.handleConn(conn)
	return nil
}

func (wt *WebsocketTransport) Handle(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// TODO: use proper origin whitelist
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println("accept:", err)
		return
	}

	conn := &Conn{
		addr: r.RemoteAddr,
		c:    c,
	}

	wt.handleConn(conn)
}

func (wt *WebsocketTransport) handleConn(conn *Conn) {
	if wt.onOpen != nil {
		wt.onOpen(conn)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		t := time.NewTimer(time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}

			start := time.Now()

			err := conn.c.Ping(ctx)
			if err != nil {
				return
			}

			end := time.Since(start)

			conn.latency.Store(end.Nanoseconds())

			t.Reset(time.Second)
		}
	}()

	go func() {
		defer cancel()

		for {
			mt, b, err := conn.c.Read(ctx)
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
	}()
}

type Conn struct {
	addr    string
	c       *websocket.Conn
	latency atomic.Int64
}

func (c *Conn) Addr() string {
	return c.addr
}

func (c *Conn) Latency() time.Duration {
	return time.Duration(c.latency.Load())
}

func (c *Conn) Write(b []byte) error {
	return c.c.Write(context.Background(), websocket.MessageText, b)
}

func NewWebsocketTransport() *WebsocketTransport {
	return &WebsocketTransport{}
}
