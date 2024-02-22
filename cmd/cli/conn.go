package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"ledctl3/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
)

type client struct {
	mux      sync.Mutex
	client   *netserver.Server[event.Event]
	addrsMux sync.Mutex
	allAddrs []net.Addr
	regAddr  string
}

func (c *client) handleEvent(addr string, e event.Event) {
	c.ProcessEvent(addr, e)
}

func newClient() (*client, error) {
	c := &client{}

	c.client = netserver.New[event.Event](-1, event.Codec)

	c.client.SetMessageHandler(func(addr string, e event.Event) {
		c.ProcessEvent(addr, e)
	})

	c.client.SetConnectHandler(func(addr string) {
		c.ProcessEvent(addr, connect{})
	})

	c.client.SetDisconnectHandler(func(addr string) {
		c.ProcessEvent(addr, disconnect{})
	})

	//fmt.Println("resolving registry address")

	mdnsResolver := mdns.NewResolver()

	addrs, err := mdnsResolver.Lookup(context.Background())
	if err != nil {
		return nil, err
	}

	go func() {
		for addr := range addrs {
			c.addrsMux.Lock()
			c.allAddrs = append(c.allAddrs, addr)
			c.addrsMux.Unlock()
		}
	}()

	c.connect()

	return c, nil
}

func (c *client) Write(e event.Event) error {
	c.mux.Lock()
	regId := c.regAddr
	c.mux.Unlock()

	if regId == "" {
		return errors.New("not connected")
	}

	return c.client.Write(regId, e)
}

func (c *client) connect() {
	c.addrsMux.Lock()
	addrs := c.allAddrs
	c.addrsMux.Unlock()

	if len(addrs) == 0 {
		time.Sleep(1 * time.Second)
		c.connect()
		return
	}

	connected := make(chan bool)

	go func() {
		var once bool

		for _, addr := range addrs {
			//fmt.Println("connecting to", addr)

			conn, err := c.client.Connect(addr)
			if err != nil {
				fmt.Println(err)
				continue
			}

			//fmt.Println("connected to", addr)

			if !once {
				once = true
				connected <- true
			}

			c.client.ProcessEvents(addr, conn)

			//fmt.Println("disconnected from", addr)

			_ = conn.Close()

			c.connect()
		}
	}()

	<-connected
}

type connect struct{}

type disconnect struct{}

func (c *client) ProcessEvent(addr string, e event.Event) {
	c.mux.Lock()
	defer c.mux.Unlock()

	switch e := e.(type) {
	case connect:
		c.handleConnect(addr)
	case disconnect:
		c.handleDisconnect()
	default:
		fmt.Println("unknown event TODO REMOVE", e)
	}
}

func (c *client) handleConnect(addr string) {
	c.regAddr = addr
}

func (c *client) handleDisconnect() {
	c.regAddr = ""
}
