package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/google/uuid"

	"ledctl3/node"
	"ledctl3/node/event"
	"ledctl3/pkg/netserver"

	//_ "ledctl3/node/drivers/display"
	_ "ledctl3/node/drivers/audio"
	//_ "ledctl3/pkg/led"
)

type Config struct {
	NodeId uuid.UUID `json:"node_id"`
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	b, err := os.ReadFile("./node.json")
	if errors.Is(err, os.ErrNotExist) {
		b, err = json.MarshalIndent(Config{
			NodeId: uuid.New(),
		}, "", "  ")
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

	res := &resolver{net: ws}
	res.start()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
