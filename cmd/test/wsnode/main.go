package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/google/uuid"

	"ledctl3/node"
	"ledctl3/node/event"
	"ledctl3/pkg/netserver"
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

	/////////////////////////////
	nod, err := node.New(
		node.Config{
			Id: cfg.NodeId,
		})
	if err != nil {
		panic(err)
	}

	/////////////////////////////

	ws := netserver.NewWebsocketTransport()

	ws.OnOpen(func(c *netserver.Conn) {
		err := nod.ProcessEvent(c, node.ConnectedEvent{})
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	ws.OnClose(func(c *netserver.Conn) {
		err := nod.ProcessEvent(c, node.DisconnectedEvent{})
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	codec := event.NewJSONCodec()
	ws.OnMessage(func(c *netserver.Conn, b []byte) {
		evt, err := codec.UnmarshalBinary(b)
		if err != nil {
			log.Println("error decoding event:", err)
			return

		}
		err = nod.ProcessEvent(c, evt)
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	err = ws.Connect("ws://localhost:1337/ws")
	if err != nil {
		fmt.Println("error processing event:", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}

func resolver() {
	// TODO
	//mdnsResolver := mdns.NewResolver()
	//
	//var allAddrs []net.Addr
	//var addrsMux sync.Mutex
	//addrs, err := mdnsResolver.Lookup(context.Background())
	//if err != nil {
	//	panic(err)
	//}
	//
	//go func() {
	//	for addr := range addrs {
	//		addrsMux.Lock()
	//		allAddrs = append(allAddrs, addr)
	//		addrsMux.Unlock()
	//	}
	//}()
	//
	//for {
	//	addrsMux.Lock()
	//	addrs := allAddrs
	//	addrsMux.Unlock()
	//
	//	for _, addr := range addrs {
	//		fmt.Println("connecting to", addr)
	//
	//		conn, err := s.Connect(addr)
	//		if err != nil {
	//			fmt.Println(err)
	//			continue
	//		}
	//
	//		s.HandleConnection(addr, conn)
	//
	//		_ = conn.Close()
	//
	//		fmt.Println("disconnected from", addr)
	//	}
	//}
}
