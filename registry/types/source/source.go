package source

import (
	"fmt"
	"net"

	"ledctl3/pkg/uuid"
)

type Source struct {
	Id   uuid.UUID
	Addr net.Addr
	Name string

	Configured bool
	Inputs     map[uuid.UUID]*Input
}

func New(id uuid.UUID, addr net.Addr) *Source {
	return &Source{
		Id:         id,
		Addr:       addr,
		Configured: false,
	}
}

func (s *Source) SetInputs(inputs map[uuid.UUID]*Input) {
	s.Inputs = inputs
	s.Configured = true
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{OutputId: %s, Name: %s}",
		s.Id, s.Name,
	)
}
