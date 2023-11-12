package main

import (
	"fmt"
	"ledctl3/event"
	"ledctl3/pkg/codec"
	"ledctl3/pkg/netclient"
	"net"
	"time"
)

func main() {
	cod := codec.NewGobCodec[event.EventIface](
		[]any{},
		map[string]any{},
		event.AssistedSetupEvent{},
		event.AssistedSetupConfigEvent{},
		event.CapabilitiesEvent{},
		event.ConnectEvent{},
		event.DataEvent{},
		event.ListCapabilitiesEvent{},
		event.SetInputConfigEvent{},
		event.SetSinkActiveEvent{},
		event.SetSourceActiveEvent{},
		event.SetSourceIdleEvent{},
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

		err := cli.Send(event.ConnectEvent{
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
