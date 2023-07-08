package source

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/pkg/event"
	"ledctl3/registry"
	regevent "ledctl3/registry/types/event"
)

type Source struct {
	id   uuid.UUID
	name string

	inputs map[uuid.UUID]*Input

	send func(event.Event) error
	recv func() <-chan event.Event
}

func NewSource(id uuid.UUID, name string, inputs map[uuid.UUID]*Input, send func(event.Event) error, recv func() <-chan event.Event) *Source {
	return &Source{
		id:   id,
		name: name,
		//state:  StateOffline,
		inputs: inputs,
		send:   send,
		recv:   recv,
	}
}

func (s *Source) Inputs() map[uuid.UUID]registry.Input {
	inputs := make(map[uuid.UUID]registry.Input)

	for id, input := range s.inputs {
		inputs[id] = input
	}

	return inputs
}

func (s *Source) Id() uuid.UUID {
	return s.id
}

func (s *Source) Name() string {
	return s.name
}

func (s *Source) Handle(e event.Event) error {
	err := s.send(e)
	if err != nil {
		return err
	}

	s.processEvent(e)
	return nil
}

func (s *Source) processEvent(e event.Event) {
	switch e := e.(type) {
	case regevent.SetSourceActiveEvent:
		//fmt.Printf("=== reg source %s: proccess SetSourceActiveEvent\n", s.id)

		for inputId := range e.Sinks {
			s.inputs[inputId].state = registry.InputStateActive
			s.inputs[inputId].sessId = e.SessionId
		}

		//fmt.Println("==== UNHANDLED PROCESS EVENT IDLE FROM SOURCE", e)
	case regevent.SetSourceIdleEvent:
		//fmt.Printf("=== reg source %s: proccess SetSourceIdleEvent\n", s.id)
		//s.state = StateIdle
		//s.sessId = uuid.Nil

		// TODO MUTATE INPUTS STATE AND SESSION

		for _, inputId := range e.InputIds {
			s.inputs[inputId].state = registry.InputStateIdle
			s.inputs[inputId].sessId = uuid.Nil
		}

		//fmt.Println("==== UNHANDLED PROCESS EVENT IDLE FROM SOURCE", e)
	default:
		fmt.Println("@@@ 2 unknown event", e, reflect.TypeOf(e))
	}
}

func (s *Source) Events() <-chan event.Event {
	return s.recv()
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{id: %s, name: %s}",
		s.id, s.name,
	)
}
