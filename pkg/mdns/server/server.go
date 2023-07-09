// Package server provides automatic device discovery on the local network
package server

import (
	"fmt"
	"os"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	server *zeroconf.Server
	name   string
	port   int
}

func New(name string, port int) (*Server, error) {
	return &Server{
		name: name,
		port: port,
	}, nil
}

func (s *Server) Start() error {
	host, err := os.Hostname()
	if err != nil {
		return err
	}

	service := fmt.Sprintf("_%s._tcp", s.name)
	zs, err := zeroconf.Register(host, service, "local", s.port, []string{"v=0.0.1"}, nil)
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
