package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"ledctl3/event"
	"ledctl3/internal/device"
	"ledctl3/internal/device/debug_output"
	screensrc "ledctl3/internal/device/screen"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

func main() {
	s := netserver2.New[event.Event](-1, event.Codec)

	dev, err := device.New(
		device.Config{
			Id: uuid.MustParse("55555555-dca5-430b-971c-fbe5b9112bfe"),
		},
		func(addr string, e event.Event) error {
			return s.Write(addr, e)
		})
	if err != nil {
		panic(err)
	}

	// 22222222-b301-47d6-b289-2a4c3327962a
	// 33333333-e72d-470e-a343-5c2cc2f1746f
	screenProv, err := screensrc.New(dev)
	if err != nil {
		panic(err)
	}

	screenProv.Start()

	out := debug_output.New(uuid.MustParse("55558888-6b50-4789-b635-16237d268efa"), 40)
	dev.AddOutput(out)

	out2 := debug_output.New(uuid.MustParse("55559999-ebd3-46dd-9d27-3d7d8443c715"), 80)
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

	var addr net.Addr
	for {
		fmt.Println("lookup")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		addr, err = mdnsResolver.Lookup(ctx)
		cancel()
		if err != nil {
			fmt.Println("error resolving: ", err)
			continue
		}
		break
	}

	for {

		//fmt.Println("@@@@@@@@@@@ CONNECTING")

		conn, dispose := s.Connect(addr)
		//fmt.Println("@@@@@@@@@@@ CONNECTED!")

		s.ProcessEvents(addr, conn)

		dispose()

		//fmt.Println("@@@@@@@@@@@@ CONNECTION INTERRUPTED")
	}
}
