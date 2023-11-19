package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"ledctl3/event"
	"ledctl3/pkg/codec"
	"ledctl3/pkg/netserver"
)

func main() {
	cod := codec.NewGobCodec[event.EventIface](
		[]any{},
		map[string]any{},
		event.AssistedSetup{},
		event.AssistedSetupConfig{},
		event.Capabilities{},
		event.Connect{},
		event.Data{},
		event.ListCapabilities{},
		event.SetInputConfig{},
		event.SetSinkActive{},
		event.SetSourceActive{},
		event.SetSourceIdle{},
	)

	var ADDR net.Addr
	var mux sync.Mutex
	cli := netserver.New[event.EventIface](1111, cod, func(addr net.Addr, e event.EventIface) {
		fmt.Println("EVENT", addr, e)
		mux.Lock()
		ADDR = addr
		mux.Unlock()
	})

	err := cli.Start()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(1 * time.Second)

		mux.Lock()
		addr := ADDR
		mux.Unlock()

		if addr == nil {
			fmt.Println("unknown addr")
			continue
		}

		err := cli.Send(addr, event.AssistedSetup{
			Event: event.Event{
				Type: event.AssistedSetup,
			},
		})
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("successfully sent event")
	}
}
