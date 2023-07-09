package source

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/event"
	"ledctl3/registry"
)

type Source struct {
	id   uuid.UUID
	name string

	inputs map[uuid.UUID]*Input

	send func(event.EventIface) error
	recv func() <-chan event.EventIface
}

func NewSource(id uuid.UUID, name string, inputs map[uuid.UUID]*Input, send func(event.EventIface) error, recv func() <-chan event.EventIface) *Source {
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

func (s *Source) Handle(e event.EventIface) error {
	err := s.send(e)
	if err != nil {
		return err
	}

	s.processEvent(e)
	return nil
}

func (s *Source) processEvent(e event.EventIface) {
	switch e := e.(type) {
	case event.SetSourceActiveEvent:
		//fmt.Printf("=== reg source %s: proccess SetSourceActiveEvent\n", s.id)

		for inputId := range e.Sinks {
			s.inputs[inputId].state = registry.InputStateActive
			s.inputs[inputId].sessId = e.SessionId
		}

		//fmt.Println("==== UNHANDLED PROCESS EVENT IDLE FROM SOURCE", e)
	case event.SetSourceIdleEvent:
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

func (s *Source) Events() <-chan event.EventIface {
	return s.recv()
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{id: %s, name: %s}",
		s.id, s.name,
	)
}
