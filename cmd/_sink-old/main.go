package main

import (
	"context"
	"fmt"
	"net"

	sinkdev "ledctl3/_sink-old"
	outputdev "ledctl3/_sink-old/debug"
	"ledctl3/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

func main() {
	snk := sinkdev.New(uuid.MustParse("d17c94aa-2fb1-4fb5-b315-f22113e8d165"))

	outputdev2a := outputdev.New(uuid.MustParse("30dc1242-2f66-4fb9-8db0-d8f29beca51c"), 20)
	outputdev2b := outputdev.New(uuid.MustParse("c715765b-29a9-42e3-aec6-f590978fb1dd"), 120)

	snk.AddOutput(outputdev2a)
	snk.AddOutput(outputdev2b)

	s := netserver2.New[event.EventIface](-1, event.Codec, func(addr net.Addr) {
		snk.ProcessEvent(addr, event.Connect{
			Event: event.Event{Type: event.Connect},
		})
	}, func(addr net.Addr, e event.EventIface) {
		snk.ProcessEvent(addr, e)
	})

	go func() {
		for msg := range snk.Messages() {
			err := s.Write(msg.Addr, msg.Event)
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
			Id: snk.Id(),
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
