// Package server provides automatic device discovery on the local network
package sourcemdns

import (
	"fmt"
	"math/rand"

	"github.com/grandcat/zeroconf"

	"ledctl3/source"
)

type Server struct {
	src         *source.Source
	server      *zeroconf.Server
	serviceName string
	port        int
}

func New(src *source.Source) (*Server, error) {
	return &Server{
		src:         src,
		serviceName: "ledctl",
		port:        1024 + rand.Intn(1024),
	}, nil
}

func (s *Server) Start() error {
	service := fmt.Sprintf("_%s._tcp", s.serviceName)
	zs, err := zeroconf.Register(s.src.Id().String(), service, "local", s.port, []string{"v=0.0.1", "uuid=" + s.src.Id().String()}, nil)
	if err != nil {
		return err
	}

	s.server = zs
	return nil
}

func (s *Server) Close() error {
	if s.server != nil {
		s.server.Shutdown()
	}

	return nil
}
