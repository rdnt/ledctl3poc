package mdns

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

type Resolver struct {
	resolver    *zeroconf.Resolver
	serviceName string
}

func NewResolver() *Resolver {
	r := &Resolver{
		serviceName: "ledctl",
	}

	for {
		zr, err := zeroconf.NewResolver(nil)
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		r.resolver = zr
		break
	}

	return r
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
			for _, ip := range e.AddrIPv4 {
				fmt.Println(ip, ip.IsPrivate())
				//var priv bool
				//for _, ipNet := range PrivateIPNetworks {
				//	if ipNet.Contains(ip) {
				//		priv = true
				//		break
				//	}
				//}
				//
				//if !priv {
				//	continue
				//}
				//
				//addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, e.Port))
				//if err != nil {
				//	fmt.Print("failed to resolve tcp address: ", err)
				//	break
				//}
				//
				//devs <- addr
			}
		}
	}(entries)

	return devs, nil
}

func (r *Resolver) Lookup(ctx context.Context) (chan net.Addr, error) {
	service := fmt.Sprintf("_%s._tcp", r.serviceName)
	entries := make(chan *zeroconf.ServiceEntry)
	addrs := make(chan net.Addr)

	err := r.resolver.Lookup(ctx, "registry", service, "local", entries)
	if err != nil {
		return nil, err
	}

	go func() {
		for e := range entries {
			for _, ip := range e.AddrIPv4 {
				if !ip.IsPrivate() {
					continue
				}

				addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, e.Port))
				if err != nil {
					fmt.Println("failed to resolve tcp address: ", err)
					continue
				}

				addrs <- addr
			}

			for _, ip := range e.AddrIPv6 {
				if !ip.IsPrivate() {
					continue
				}

				addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, e.Port))
				if err != nil {
					fmt.Println("failed to resolve tcp6 address: ", err)
					continue
				}

				addrs <- addr
			}
		}
	}()

	return addrs, nil
}
