package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"

	"ledctl3/event"
	"ledctl3/pkg/netbroker"
)

func init() {
	gob.Register([]any{})
	gob.Register(map[string]any{})
	gob.Register(event.AssistedSetup{})
	gob.Register(event.AssistedSetupConfig{})
	gob.Register(event.Capabilities{})
	gob.Register(event.Connect{})
	gob.Register(event.Data{})
	gob.Register(event.ListCapabilities{})
	gob.Register(event.SetInputConfig{})
	gob.Register(event.SetSinkActive{})
	gob.Register(event.SetSourceActive{})
	gob.Register(event.SetSourceIdle{})
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
	br.Start(":1111")

	br.Receive(func(addr net.Addr, e event.EventIface) {
		fmt.Println("RECEIVED EVENT")
		fmt.Println(e)

		err := br.Send(addr, event.ListCapabilities{
			Event: event.Event{
				Type: event.ListCapabilities,
			},
		})
		fmt.Println("SEND LIST CAPS", err)
	})

	//ip := net.ParseIP("127.0.0.1")
	//addr := &net.TCPAddr{
	//	IP:   ip,
	//	Port: 8080,
	//	Zone: "",
	//}
	//
	//br.AddServer(addr)

	select {}
}
