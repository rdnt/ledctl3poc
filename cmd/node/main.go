package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"ledctl3/event"
	"ledctl3/node"
	_ "ledctl3/node/screen"
	_ "ledctl3/pkg/led"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
	"ledctl3/pkg/uuid"
)

type Config struct {
	NodeId uuid.UUID `json:"node_id"`
}

func main() {
	b, err := os.ReadFile("./node.json")
	if errors.Is(err, os.ErrNotExist) {
		b, err = json.Marshal(Config{
			NodeId: uuid.New(),
		})
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("./node.json", b, 0644)
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		panic(err)
	}

	s := netserver.New[event.Event](-1, event.Codec)

	dev, err := node.New(
		node.Config{
			Id: cfg.NodeId,
		},
		func(addr string, e event.Event) error {
			return s.Write(addr, e)
		})
	if err != nil {
		panic(err)
	}

	// 22222222-b301-47d6-b289-2a4c3327962a
	// 33333333-e72d-470e-a343-5c2cc2f1746f
	//if runtime.GOOS == "windows" {
	//	screenProv, err := screensrc.New(dev, "dxgi")
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	screenProv.Start()
	//}

	s.SetMessageHandler(func(addr string, e event.Event) {
		dev.ProcessEvent(addr, e)
	})

	s.SetConnectHandler(func(addr string) {
		//fmt.Println("CONNECT CALLED")
		dev.ProcessEvent(addr, node.ConnectedEvent{})
	})

	s.SetDisconnectHandler(func(addr string) {
		//fmt.Println("DISCONNECT CALLED")
		dev.ProcessEvent(addr, node.DisconnectedEvent{})
	})

	fmt.Println(cfg.NodeId, "started")

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

			s.HandleConnection(addr, conn)

			_ = conn.Close()

			fmt.Println("disconnected from", addr)
		}
	}
}
