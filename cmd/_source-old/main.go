package main

import (
	"context"
	"fmt"
	"net"

	sourcedev "ledctl3/_source-old"
	screensrc "ledctl3/_source-old/screen"
	"ledctl3/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

func main() {
	src, err := sourcedev.New(uuid.MustParse("4282186d-dca5-430b-971c-fbe5b9112bfe"))
	handle(err)

	screenProv, err := screensrc.New(src)
	handle(err)

	screenProv.Start()

	s := netserver2.New[event.EventIface](-1, event.Codec, func(addr net.Addr) {
		src.ProcessEvent(addr, event.Connect{
			Event: event.Event{Type: event.Connect},
		})
	}, func(addr net.Addr, e event.EventIface) {
		src.ProcessEvent(addr, e)
	})

	go func() {
		for msg := range src.Messages() {
			err = s.Write(msg.Addr, msg.Event)
			if err != nil {
				fmt.Print("error sending event: ", err)
			}
		}
	}()

	mdnsResolver, err := mdns.NewResolver()
	handle(err)

	for {
		addr, err := mdnsResolver.Lookup(context.Background())
		if err != nil {
			fmt.Println("error resolving: ", err)
			continue
		}

		//fmt.Println("@@@@@@@@@@@@ CONNECTING")
		conn, dispose := s.Connect(addr)

		err = s.Write(addr, event.Connect{
			Id: src.Id(),
		})
		if err != nil {
			fmt.Println(err)
			dispose()
		}

		s.ProcessEvents(addr, conn)

		//fmt.Println("@@@@@@@@@@@@ CONNECTION INTERRUPTED")
	}

	fmt.Println("idle")
	select {}
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
