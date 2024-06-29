package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"ledctl3/cmd/registry/state"
	"ledctl3/node/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
	"ledctl3/registry"
)

func main() {
	sh := state.NewHolder()
	reg := registry.New(sh)

	ws := netserver.NewWebsocketTransport()

	ws.OnOpen(func(c *netserver.Conn) {
		err := reg.ProcessEvent(c, registry.ConnectedEvent{})
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	ws.OnClose(func(c *netserver.Conn) {
		err := reg.ProcessEvent(c, registry.DisconnectedEvent{})
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
		err = reg.ProcessEvent(c, evt)
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	// ROUTING

	r := chi.NewRouter()

	r.Get("/ws", ws.Handle)

	go func() {
		err := http.ListenAndServe(":1337", r)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go broadcastD()

	go func() {
		for {
			time.Sleep(3 * time.Second)

			err := reg.EnableProfile(uuid.MustParse("aaaaaaaa-0000-4000-0000-000000000000"))
			if err != nil {
				fmt.Println("error enabling profile:", err)
			}
			break
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	sh.Stop()
}

func broadcastD() {
	for {
		err := broadcast()
		if err != nil {
			fmt.Println("error broadcasting:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return
	}
}

func broadcast() error {
	mdnsServer, err := mdns.NewServer("registry", 1337)
	if err != nil {
		return err
	}

	err = mdnsServer.Start()
	if err != nil {
		return err
	}

	return nil
}
