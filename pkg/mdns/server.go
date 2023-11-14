package mdns

import (
	"fmt"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	server      *zeroconf.Server
	serviceName string
	instance    string
	port        int
}

func NewServer(instance string, port int) (*Server, error) {
	return &Server{
		instance:    instance,
		serviceName: "ledctl",
		port:        port,
	}, nil
}

func (s *Server) Start() error {
	service := fmt.Sprintf("_%s._tcp", s.serviceName)
	zs, err := zeroconf.Register(s.instance, service, "local", s.port, []string{"v=0.0.1"}, nil)
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
