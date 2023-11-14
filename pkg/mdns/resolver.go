package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/grandcat/zeroconf"
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
	Addr net.Addr
}

func (r *Resolver) Browse(ctx context.Context) (<-chan net.Addr, error) {
	devs := make(chan net.Addr)
	entries := make(chan *zeroconf.ServiceEntry, 10)

	service := fmt.Sprintf("_%s._tcp", r.serviceName)

	err := r.resolver.Browse(ctx, service, "local", entries)
	if err != nil {
		return nil, err
	}

	go func(entries chan *zeroconf.ServiceEntry) {
		for e := range entries {
			addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", e.AddrIPv4[0], e.Port))
			if err != nil {
				fmt.Print("failed to resolve tcp address: ", err)
				break
			}

			devs <- addr
		}
	}(entries)

	return devs, nil
}

func (r *Resolver) Lookup(ctx context.Context) (net.Addr, error) {
	service := fmt.Sprintf("_%s._tcp", r.serviceName)
	entries := make(chan *zeroconf.ServiceEntry)

	err := r.resolver.Lookup(ctx, "registry", service, "local", entries)
	if err != nil {
		return nil, err
	}

	e := <-entries

	if e == nil {
		return nil, errors.New("no device found")
	}

	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", e.AddrIPv4[0], e.Port))
	if err != nil {
		fmt.Print("failed to resolve tcp address: ", err)
		return nil, err
	}

	return addr, nil
}
