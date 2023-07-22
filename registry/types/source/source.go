package source

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Source struct {
	Id   uuid.UUID
	Name string

	Configured bool
	Inputs     map[uuid.UUID]*Input
}

func New(id uuid.UUID) *Source {
	return &Source{
		Id:         id,
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
