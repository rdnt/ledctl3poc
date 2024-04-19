package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"

	"ledctl3/cmd/registry/state"
	"ledctl3/node/event"
	"ledctl3/orchestrator"
	"ledctl3/pkg/netserver"
)

func main() {
	sh := state.NewHolder()
	reg := orchestrator.New(sh)

	ws := netserver.NewWebsocketTransport()

	ws.OnOpen(func(c *netserver.Conn) {
		err := reg.ProcessEvent(c, orchestrator.ConnectedEvent{})
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	ws.OnClose(func(c *netserver.Conn) {
		err := reg.ProcessEvent(c, orchestrator.DisconnectedEvent{})
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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
