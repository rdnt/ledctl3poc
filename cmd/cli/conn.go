package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"ledctl3/node/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
)

type client struct {
	mux      sync.Mutex
	client   *netserver.Server[event.Event, event.Event]
	addrsMux sync.Mutex
	allAddrs []net.Addr
	regAddr  string
}

func (c *client) handleEvent(addr string, e event.Event) {
	c.ProcessEvent(addr, e)
}

func newClient() (*client, error) {
	allAddrs := make([]net.Addr, 0)
	localAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1337")
	if err == nil {
		allAddrs = append(allAddrs, localAddr)
	}
	c := &client{allAddrs: allAddrs}

	c.client = netserver.New[event.Event, event.Event](-1, event.JSONCodec{}, event.JSONCodec{})

	c.client.SetMessageHandler(func(addr string, e event.Event) {
		c.ProcessEvent(addr, e)
	})

	c.client.SetConnectHandler(func(addr string) {
		c.handleConnect(addr)
	})

	c.client.SetDisconnectHandler(func(addr string) {
		c.handleDisconnect()
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

func (c *client) Request(e event.Event) error {
	c.mux.Lock()
	regId := c.regAddr
	c.mux.Unlock()

	if regId == "" {
		return errors.New("not connected")
	}

	return c.client.Request(regId, e)
}

var once bool

var connected = make(chan bool)

func (c *client) connect() {
	//fmt.Println("connect")
	c.addrsMux.Lock()
	addrs := c.allAddrs
	//fmt.Println(addrs)
	c.addrsMux.Unlock()

	go func() {

		for _, addr := range addrs {
			//fmt.Println("connecting to", addr)

			conn, err := c.client.Connect(addr)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					//fmt.Println("error during read: ", err)
				}
				continue
			}

			//fmt.Println("connected to", addr)

			if !once {
				once = true
				connected <- true
			}

			c.client.HandleConnection(addr, conn)

			//fmt.Println("disconnected from", addr)

			_ = conn.Close()
			break
		}

		time.Sleep(10 * time.Millisecond)
		c.connect()
	}()

	<-connected
}

type connect struct{}

type disconnect struct{}

func (c *client) ProcessEvent(addr string, e event.Event) {
	c.mux.Lock()
	defer c.mux.Unlock()

	switch e := e.(type) {
	//case connect:
	//	c.handleConnect(addr)
	//case disconnect:
	//	c.handleDisconnect()
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
