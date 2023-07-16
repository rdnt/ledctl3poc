package source

import (
	"fmt"

	"github.com/google/uuid"
)

type Source struct {
	Id   uuid.UUID
	Name string

	Inputs map[uuid.UUID]*Input
}

func NewSource(id uuid.UUID, name string, inputs map[uuid.UUID]*Input) *Source {
	return &Source{
		Id:     id,
		Name:   name,
		Inputs: inputs,
	}
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{OutputId: %s, Name: %s}",
		s.Id, s.Name,
	)
}
