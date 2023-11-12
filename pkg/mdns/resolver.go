package mdns

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"ledctl3/pkg/uuid"
	"net"
)

type Resolver struct {
	resolver    *zeroconf.Resolver
	serviceName string
}

func NewResolver() (*Resolver, error) {
	zr, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		resolver:    zr,
		serviceName: "ledctl",
	}, nil
}

type OnRegistryFound func(addr net.Addr)

type Device struct {
	Id   uuid.UUID
	Addr net.Addr
}

func (r *Resolver) Browse(ctx context.Context) (<-chan Device, error) {
	devs := make(chan Device)
	entries := make(chan *zeroconf.ServiceEntry, 10)

	service := fmt.Sprintf("_%s._tcp", r.serviceName)

	err := r.resolver.Browse(ctx, service, "local", entries)
	if err != nil {
		return nil, err
	}

	go func(entries chan *zeroconf.ServiceEntry) {
		for e := range entries {
			id, err := uuid.Parse(e.Instance)
			if err != nil {
				fmt.Print("failed to parse uuid: ", err)
				break
			}

			fmt.Println("@@@", e.AddrIPv4[0])

			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", e.AddrIPv4[0], e.Port))
			if err != nil {
				fmt.Print("failed to resolve tcp address: ", err)
				break
			}

			fmt.Println("@@@", addr)

			devs <- Device{
				Id:   id,
				Addr: addr,
			}
		}
	}(entries)

	return devs, nil
}
