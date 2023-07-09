package resolver

import (
	"context"
	"fmt"
	"net"

	"github.com/grandcat/zeroconf"
)

type Resolver struct {
	resolver *zeroconf.Resolver
	name     string
	port     int
}

type Device struct {
	Name string
	Port int
	IPv4 []net.IP
	IPv6 []net.IP
}

func New(name string, port int) (*Resolver, error) {
	zr, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		resolver: zr,
		name:     name,
		port:     port,
	}, nil
}

func (r *Resolver) Browse(ctx context.Context) (chan *Device, error) {
	entries := make(chan *zeroconf.ServiceEntry)
	devices := make(chan *Device)

	service := fmt.Sprintf("_%s._tcp", r.name)

	go func(entries chan *zeroconf.ServiceEntry) {
		for e := range entries {
			if e.Port != r.port {
				continue
			}

			d := &Device{
				Name: e.Instance,
				Port: e.Port,
				IPv4: e.AddrIPv4,
				IPv6: e.AddrIPv6,
			}

			devices <- d
		}

		close(devices)
	}(entries)

	err := r.resolver.Browse(ctx, service, "local", entries)
	if err != nil {
		return nil, err
	}

	return devices, nil
}
