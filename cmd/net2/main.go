package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"ledctl3/event"
	"ledctl3/pkg/netbroker"
	"net"
	"time"
)

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(event.AssistedSetupEvent{})
	gob.Register(event.AssistedSetupConfigEvent{})
	gob.Register(event.CapabilitiesEvent{})
	gob.Register(event.ConnectEvent{})
	gob.Register(event.DataEvent{})
	gob.Register(event.ListCapabilitiesEvent{})
	gob.Register(event.SetInputConfigEvent{})
	gob.Register(event.SetSinkActiveEvent{})
	gob.Register(event.SetSourceActiveEvent{})
	gob.Register(event.SetSourceIdleEvent{})
}

func main() {
	br := netbroker.New[event.EventIface](func(e event.EventIface) ([]byte, error) {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(&e)
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}, func(b []byte) (event.EventIface, error) {
		r := bytes.NewReader(b)

		dec := gob.NewDecoder(r)
		var e event.EventIface
		err := dec.Decode(&e)
		if err != nil {
			return nil, err
		}

		return e, nil
	})
	//br.Start(":2222")

	ip := net.ParseIP("127.0.0.1")
	addr := &net.TCPAddr{
		IP:   ip,
		Port: 1111,
		Zone: "",
	}

	br.AddServer(addr)

	time.Sleep(1 * time.Second)

	err := br.Send(addr, event.ConnectEvent{
		Event: event.Event{
			Type: event.Connect,
		},
		Id: "my pretty id :D",
	})
	fmt.Print("send err: ", err, "\n")

	select {}
}
