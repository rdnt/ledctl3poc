package main

import (
	"fmt"
	"net"
	"time"

	"ledctl3/event"
	"ledctl3/pkg/codec"
	"ledctl3/pkg/netclient"
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

	ip := net.ParseIP("127.0.0.1")
	addr := &net.TCPAddr{
		IP:   ip,
		Port: 1111,
		Zone: "",
	}

	cli := netclient.New[event.EventIface](addr, cod, func(e event.EventIface) {
		fmt.Println("EVENT", e)
	})

	for {
		time.Sleep(1 * time.Second)

		err := cli.Send(event.Connect{
			Event: event.Event{
				Type: event.Connect,
			},
		})
		if err != nil {
			fmt.Println("error sending event: ", err)
			continue
		}

		fmt.Println("successfully sent event")

	}
}
