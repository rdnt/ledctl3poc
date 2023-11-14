package source

import (
	"fmt"
	"net"

	"ledctl3/pkg/uuid"
)

type Source struct {
	Id     uuid.UUID
	Addr   net.Addr
	Name   string
	Inputs map[uuid.UUID]*Input
}

func New(id uuid.UUID, addr net.Addr) *Source {
	return &Source{
		Id:   id,
		Addr: addr,
	}
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{Id: %s, Name: %s}",
		s.Id, s.Name,
	)
}
