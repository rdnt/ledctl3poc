package source

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"ledctl3/event"
)

type Source struct {
	id   uuid.UUID
	name string

	inputs map[uuid.UUID]*Input
	schema map[string]any
	cfg    map[string]any
}

func NewSource(id uuid.UUID, name string, inputs map[uuid.UUID]*Input) *Source {
	return &Source{
		id:     id,
		name:   name,
		inputs: inputs,
	}
}

func (s *Source) Inputs() map[uuid.UUID]*Input {
	inputs := make(map[uuid.UUID]*Input)

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

func (s *Source) Process(e event.EventIface) error {
	switch e := e.(type) {
	case event.SetSourceActiveEvent:
		//fmt.Printf("=== reg source %s: proccess SetSourceActiveEvent\n", s.id)

		for _, input := range e.Inputs {
			s.inputs[input.Id].state = InputStateActive
			s.inputs[input.Id].sessId = e.SessionId

			s.inputs[input.Id].sinks = lo.Map(input.Sinks, func(sink event.SetSourceActiveEventSink, index int) sinkConfig {
				return sinkConfig{
					Id: sink.Id,
					Outputs: lo.Map(sink.Outputs, func(output event.SetSourceActiveEventOutput, _ int) outputConfig {
						return outputConfig(output)
					}),
				}
			})

			//
			//var sinks []sink
			//for _, cfg := range input.Sinks {
			//	s := sink{
			//		id: cfg.Id,
			//	}
			//
			//	for _, out := range cfg.Outputs {
			//		s.outputs = append(s.outputs, output{
			//			id:   out.Id,
			//			leds: out.Leds,
			//		})
			//	}
			//
			//	sinks = append(sinks, s)
			//}
		}

		return nil

		//fmt.Println("==== UNHANDLED PROCESS EVENT IDLE FROM SOURCE", e)
	case event.SetSourceIdleEvent:
		//fmt.Printf("=== reg source %s: proccess SetSourceIdleEvent\n", s.id)
		//s.state = StateIdle
		//s.sessId = uuid.Nil

		// TODO MUTATE INPUTS STATE AND SESSION

		for _, inputId := range e.InputIds {
			s.inputs[inputId].state = InputStateIdle
			s.inputs[inputId].sessId = uuid.Nil
		}

		return nil

	//fmt.Println("==== UNHANDLED PROCESS EVENT IDLE FROM SOURCE", e)
	case event.SetInputConfigEvent:
		input, ok := s.inputs[e.InputId]
		if !ok {
			return fmt.Errorf("input %s not found", e.InputId)
		}

		err := input.ApplyConfig(e.Config)
		if err != nil {
			return fmt.Errorf("apply input config: %w", err)
		}

		return nil
	}

	fmt.Println("@@@ 2 unknown event", e, reflect.TypeOf(e))

	return nil
}

func (s *Source) String() string {
	return fmt.Sprintf(
		"source{id: %s, name: %s}",
		s.id, s.name,
	)
}
