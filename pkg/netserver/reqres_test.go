package netserver_test

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"ledctl3/pkg/netserver"
)

type codec struct{}

func (c codec) MarshalEvent(e any) ([]byte, error) {
	return json.Marshal(e)
}

func (c codec) UnmarshalEvent(b []byte, e *any) error {
	return json.Unmarshal(b, e)
}

func TestReqResServer(t *testing.T) {

	s := netserver.New[any](1234, codec{})

	s.SetRequestHandler(func(addr string, e any, respond func(any) error) {
		t.Log("[server] received request", addr, e)

		resp, err := s.Request(addr, "i the server is requesting additional from client")
		if err != nil {
			t.Fatal(err)
		}

		t.Log("[server] received response", addr, resp)

		err = respond(fmt.Sprint("i the server received ", e, " and responded with ", resp))
		if err != nil {
			t.Fatal(err)
		}
	})

	err := s.Start()
	if err != nil {
		t.Fatal(err)
	}

	select {}
}

func TestReqResClient(t *testing.T) {
	c := netserver.New[any](-1, codec{})

	c.SetRequestHandler(func(addr string, e any, respond func(any) error) {
		t.Log("[client] received request", addr, e)

		resp, err := c.Request(addr, ":D:D")
		if err != nil {
			t.Fatal(err)
		}

		t.Log("[client] received response", addr, resp)

		err = respond(fmt.Sprint("i the client received ", e, " and responded with ", resp))
		if err != nil {
			t.Fatal(err)
		}
	})

	addr, err := net.ResolveTCPAddr("tcp", "localhost:1234")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := c.Connect(addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err = c.HandleConnection(conn.RemoteAddr(), conn)
		if err != nil {
			t.Fatal(err)
		}
	}()

	res, err := c.Request(conn.RemoteAddr().String(), "hello world :)")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("response (todo)", res)

	select {}
}
