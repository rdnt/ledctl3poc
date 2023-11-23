package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"ledctl3/event"
	"ledctl3/internal/device"
	"ledctl3/internal/device/debug_output"
	screensrc "ledctl3/internal/device/screen"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

type Config struct {
	DeviceId  uuid.UUID `json:"device_id"`
	Output1Id uuid.UUID `json:"output1_id"`
	Output2Id uuid.UUID `json:"output2_id"`
}

func main() {
	b, err := os.ReadFile("../device.json")
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}

	s := netserver2.New[event.Event](-1, event.Codec)

	dev, err := device.New(
		device.Config{
			Id: cfg.DeviceId,
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

	out := debug_output.New(cfg.Output1Id, 40)
	dev.AddOutput(out)

	out2 := debug_output.New(cfg.Output2Id, 80)
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

	fmt.Println(cfg.DeviceId, "started")

	addr, err := mdnsResolver.Lookup(context.Background())
	if err != nil {
		panic(err)
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
