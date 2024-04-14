package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type websocketTransport struct {
	connsMux sync.Mutex
	conns    map[uuid.UUID]*connection
}

type connection struct {
	*websocket.Conn
	reqsMux sync.Mutex
	reqs    map[uuid.UUID]chan []byte
}

type request struct {
	RequestID uuid.UUID
	Request   json.RawMessage
}

func (wt *websocketTransport) Request(ctx context.Context, id uuid.UUID, b []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	wt.connsMux.Lock()
	conn, ok := wt.conns[id]
	wt.connsMux.Unlock()
	if !ok {
		return nil, errors.New("destination not connected")
	}

	//reqId := uuid.New()
	reqId := uuid.Nil

	b, err := json.Marshal(request{
		RequestID: reqId,
		Request:   b,
	})
	if err != nil {
		return nil, err
	}

	respCh := make(chan []byte)

	wt.connsMux.Lock()
	conn.reqs[reqId] = respCh
	wt.connsMux.Unlock()

	defer func() {
		wt.connsMux.Lock()
		delete(conn.reqs, reqId)
		wt.connsMux.Unlock()
	}()

	err = conn.Conn.Write(ctx, websocket.MessageText, b)
	if err != nil {
		return nil, err
	}

	fmt.Println("wait")
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case b := <-respCh:
		return b, nil
	}
}

func (wt *websocketTransport) wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("accept:", err)
		return
	}

	mt, b, err := c.Read(context.Background())
	if err != nil {
		log.Println("read:", err)
		return
	}

	if mt != websocket.MessageText {
		return
	}

	id, err := uuid.Parse(string(b))
	if err != nil {
		log.Println("parse uuid:", err)
		return
	}

	wt.connsMux.Lock()
	conn := &connection{
		Conn: c,
		reqs: make(map[uuid.UUID]chan []byte),
	}
	wt.conns[id] = conn
	wt.connsMux.Unlock()

	fmt.Println("new conn", id)

	wt.handleConn(conn)

	//err = c.Write(ctx, mt, data)
	//if err != nil {
	//	log.Println("write:", err)
	//	break
	//}
	//atomic.AddUint64(&qps, 1)
}

type response struct {
	ResponseID uuid.UUID
	Response   json.RawMessage
}

func (wt *websocketTransport) handleConn(c *connection) {
	for {
		mt, b, err := c.Read(context.Background())
		if err != nil {
			log.Println("read:", err)
			return
		}

		if mt != websocket.MessageText {
			continue
		}

		var resp response
		err = json.Unmarshal(b, &resp)
		if err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		wt.connsMux.Lock()
		req, ok := c.reqs[resp.ResponseID]
		if !ok {
			wt.connsMux.Unlock()
			continue
		}

		req <- resp.Response
		close(req)

		wt.connsMux.Unlock()
	}
}

func main() {
	r := chi.NewRouter()

	wt := &websocketTransport{
		conns: make(map[uuid.UUID]*connection),
	}

	r.Get("/ws", wt.wsHandler)

	go func() {
		err := http.ListenAndServe(":1337", r)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go client()

	time.Sleep(2 * time.Second)

	fmt.Println("sent", "hi:D")
	resp, err := wt.Request(context.Background(), uuid.Nil, []byte("\"hi:D\""))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("got resp!", string(resp))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func client() {
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:1337/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

	err = c.Write(ctx, websocket.MessageText, []byte(`00000000-0000-0000-0000-000000000000`))
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	err = c.Write(ctx, websocket.MessageText, []byte(`{"ResponseID": "00000000-0000-0000-0000-000000000000", "Response": "henlo:D"}`))
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	c.Close(websocket.StatusNormalClosure, "")
}
