package main

import (
	"fmt"
	"github.com/phayes/freeport"
	"ledctl3/event"
	"ledctl3/pkg/codec"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
	sourcedev "ledctl3/source"
	screensrc "ledctl3/source/screen"
	"log"
	"net"
)

func main() {
	src, err := sourcedev.New(uuid.MustParse("4282186d-dca5-430b-971c-fbe5b9112bfe"))
	handle(err)

	screenProv, err := screensrc.New(src)
	handle(err)

	screenProv.Start()

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

	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}

	s := netserver2.New[event.EventIface](port, cod, func(addr net.Addr, e event.EventIface) {
		src.ProcessEvent(addr, e)
	})

	err = s.Start()
	handle(err)

	go func() {
		for msg := range src.Messages() {
			err = s.Write(msg.Addr, msg.Event)
			if err != nil {
				fmt.Print("error sending event: ", err)
			}
		}
	}()

	mdnsServer, err := mdns.NewServer(src.Id().String(), port)
	handle(err)

	err = mdnsServer.Start()
	handle(err)

	select {}
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
