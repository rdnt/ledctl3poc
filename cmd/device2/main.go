package main

import (
	"context"
	"fmt"

	"ledctl3/event"
	"ledctl3/internal/device"
	"ledctl3/internal/device/debug_output"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

func main() {
	s := netserver2.New[event.Event](-1, event.Codec)

	dev, err := device.New(
		device.Config{
			Id: uuid.MustParse("66666666-dca5-430b-971c-fbe5b9112bfe"),
		},
		func(addr string, e event.Event) error {
			return s.Write(addr, e)
		})
	if err != nil {
		panic(err)
	}

	out := debug_output.New(uuid.MustParse("88888888-6b50-4789-b635-16237d268efa"), 40)
	dev.AddOutput(out)

	out2 := debug_output.New(uuid.MustParse("99999999-ebd3-46dd-9d27-3d7d8443c715"), 80)
	dev.AddOutput(out2)

	s.SetMessageHandler(func(addr string, e event.Event) {
		dev.ProcessEvent(addr, e)
	})

	s.SetConnectHandler(func(addr string) {
		//fmt.Println("CONNECT CALLED")
		dev.ProcessEvent(addr, event.Connect{})
	})

	s.SetDisconnectHandler(func(addr string) {
		//fmt.Println("DISCONNECT CALLED")
		dev.ProcessEvent(addr, event.Disconnect{})
	})

	mdnsResolver, err := mdns.NewResolver()
	if err != nil {
		panic(err)
	}

	for {
		addr, err := mdnsResolver.Lookup(context.Background())
		if err != nil {
			fmt.Println("error resolving: ", err)
			continue
		}

		//fmt.Println("@@@@@@@@@@@ CONNECTING")

		conn, dispose := s.Connect(addr)
		//fmt.Println("@@@@@@@@@@@ CONNECTED!")

		s.ProcessEvents(addr, conn)

		dispose()

		//fmt.Println("@@@@@@@@@@@@ CONNECTION INTERRUPTED")
	}

	fmt.Println("idle")
	select {}
}
