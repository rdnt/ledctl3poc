package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"ledctl3/event"
	"ledctl3/internal/device"
	"ledctl3/internal/device/debug_output"
	"ledctl3/internal/device/debug_output_ws"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
	"ledctl3/pkg/uuid"
	"ledctl3/pkg/ws281x"
)

type Config struct {
	DeviceId  uuid.UUID `json:"device_id"`
	Output1Id uuid.UUID `json:"output1_id"`
	Output2Id uuid.UUID `json:"output2_id"`
	Output3Id uuid.UUID `json:"output3_id"`
}

func main() {
	fmt.Println("starting")

	b, err := os.ReadFile("../device.json")
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}

	s := netserver.New[event.Event](-1, event.Codec)

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

	out1 := debug_output.New(cfg.Output1Id, 40)
	dev.AddOutput(out1)

	out2 := debug_output.New(cfg.Output2Id, 80)
	dev.AddOutput(out2)

	engine, err := ws281x.Init(18, 3, 255, "grb")
	if err != nil {
		panic(err)
	}

	out3 := debug_output_ws.New(cfg.Output3Id, 3, engine)
	dev.AddOutput(out3)

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

	fmt.Println(cfg.DeviceId, "started")

	fmt.Println("resolving registry address")

	mdnsResolver := mdns.NewResolver()

	var allAddrs []net.Addr
	var addrsMux sync.Mutex
	addrs, err := mdnsResolver.Lookup(context.Background())
	if err != nil {
		panic(err)
	}

	go func() {
		for addr := range addrs {
			addrsMux.Lock()
			allAddrs = append(allAddrs, addr)
			addrsMux.Unlock()
		}
	}()

	for {
		addrsMux.Lock()
		addrs := allAddrs
		addrsMux.Unlock()

		for _, addr := range addrs {
			fmt.Println("connecting to", addr)

			conn, err := s.Connect(addr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			s.ProcessEvents(addr, conn)

			_ = conn.Close()

			fmt.Println("disconnected from", addr)
		}
	}
}
