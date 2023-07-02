package types

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"time"
)

type Source struct {
	id      uuid.UUID
	name    string
	address net.Addr
	state   State
	events  chan Event
}

func (s *Source) Id() uuid.UUID {
	return s.id
}

func (s *Source) Name() string {
	return s.name
}

func (s *Source) Address() net.Addr {
	return s.address
}

func (s *Source) State() State {
	return s.state
}

func (s *Source) SetState(state State) {
	s.state = state

	go func() {
		for {
			time.Sleep(1 * time.Second)
			s.events <- Event{}
		}
	}()
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"src{id: %s, name: %s, address: %s, state: %s}",
		s.id, s.name, s.address, s.state,
	)
}

func (s *Source) Events() chan Event {
	return s.events
}

func NewSource(name string, address net.Addr) *Source {
	return &Source{
		id:      uuid.New(),
		name:    name,
		address: address,
		state:   StateOffline,
		events:  make(chan Event),
	}
}
