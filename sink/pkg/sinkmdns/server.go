// Package server provides automatic device discovery on the local network
package sinkmdns

import (
	"fmt"
	"math/rand"

	"github.com/grandcat/zeroconf"

	"ledctl3/sink"
)

type Server struct {
	sink        *sink.Sink
	server      *zeroconf.Server
	serviceName string
	port        int
}

func New(snk *sink.Sink) (*Server, error) {
	return &Server{
		sink:        snk,
		serviceName: "ledctl",
		port:        1024 + rand.Intn(1024),
	}, nil
}

func (s *Server) Start() error {
	service := fmt.Sprintf("_%s._tcp", s.serviceName)
	zs, err := zeroconf.Register(s.sink.Id().String(), service, "local", s.port, []string{"v=0.0.1", "uuid=" + s.sink.Id().String()}, nil)
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
