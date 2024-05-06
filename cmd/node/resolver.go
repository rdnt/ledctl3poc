package main

import (
	"context"
	"sync"

	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
)

type resolver struct {
	net   *netserver.WebsocketTransport
	mux   sync.Mutex
	addrs []string
}

func (r *resolver) start() {
	go r.discoverD()
}

func (r *resolver) connectD() {
	for {
		r.mux.Lock()
		addrs := r.addrs
		r.mux.Unlock()

		for _, addr := range addrs {
			err := r.net.Connect(addr)
			if err != nil {
				//fmt.Println(err)
			}
		}
	}
}

func (r *resolver) discoverD() {
	for {
		r.mux.Lock()
		r.addrs = nil
		r.mux.Unlock()

		mdnsResolver := mdns.NewResolver()

		addrs, err := mdnsResolver.Lookup(context.Background())
		if err != nil {
			continue
		}

		for addr := range addrs {
			r.mux.Lock()

			r.addrs = append(r.addrs, "ws://"+addr.String()+"/ws")
			if len(r.addrs) == 1 {
				go r.connectD()
			}

			r.mux.Unlock()
		}
	}
}
