package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/grandcat/zeroconf"

	"ledctl3/registry"
)

type Resolver struct {
	reg         *registry.Registry
	resolver    *zeroconf.Resolver
	serviceName string
}

func New(reg *registry.Registry) (*Resolver, error) {
	zr, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		reg:         reg,
		resolver:    zr,
		serviceName: "ledctl",
	}, nil
}

func (r *Resolver) Browse(ctx context.Context) error {
	entries := make(chan *zeroconf.ServiceEntry, 10)

	service := fmt.Sprintf("_%s._tcp", r.serviceName)

	err := r.resolver.Browse(ctx, service, "local", entries)
	if err != nil {
		return err
	}

	go func(entries chan *zeroconf.ServiceEntry) {
		for e := range entries {
			for _, txt := range e.Text {
				if strings.HasPrefix(txt, "uuid=") {
					//var err error
					//id, err := uuid.Parse(strings.TrimPrefix(txt, "uuid="))
					//if err != nil {
					//	fmt.Print("failed to parse uuid: ", err)
					//	break
					//}
					//
					//addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", e.AddrIPv4[0], e.Port))
					//if err != nil {
					//	fmt.Print("failed to resolve tcp address: ", err)
					//	break
					//}

					//err = r.reg.RegisterDevice(id, addr)
					//if errors.Is(err, registry.ErrDeviceExists) {
					//	break
					//} else if err != nil {
					//	fmt.Print("failed to add source: ", err)
					//	break
					//}
				}
			}
		}
	}(entries)

	return nil
}
