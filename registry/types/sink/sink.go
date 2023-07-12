package sink

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/event"
	regevent "ledctl3/event"
)

type Calibration struct {
	R float64
	G float64
	B float64
	A float64
}

type Sink struct {
	id   uuid.UUID
	name string

	outputs map[uuid.UUID]*Output
}

//type State string
//
//const (
//	StateOffline State = "offline"
//	StateIdle    State = "idle"
//	StateActive  State = "active"
//)

func NewSink(id uuid.UUID, name string, outputs map[uuid.UUID]*Output) *Sink {
	return &Sink{
		id:      id,
		name:    name,
		outputs: outputs,
	}
}

func (s *Sink) Id() uuid.UUID {
	return s.id
}

func (s *Sink) Name() string {
	return s.name
}

func (s *Sink) Leds() int {
	var leds int
	for _, dev := range s.outputs {
		leds += dev.leds
	}

	return leds
}

func (s *Sink) Calibration() map[int]Calibration {
	calib := make(map[int]Calibration)

	var acc int
	for _, out := range s.outputs {
		for i, c := range out.calibration {
			calib[i+acc] = c
		}

		acc += out.leds
	}

	return calib
}

func (s *Sink) String() string {
	return fmt.Sprintf(
		"sink{id: %s, name: %s, leds: %d, calibration: %v}",
		s.id, s.name, s.Leds(), s.Calibration(),
	)
}

func (s *Sink) Outputs() map[uuid.UUID]*Output {
	outputs := make(map[uuid.UUID]*Output)

	for id, output := range s.outputs {
		outputs[id] = output
	}

	return outputs
}

func (s *Sink) Process(e event.EventIface) {
	switch e := e.(type) {
	case regevent.SetSinkActiveEvent:
		for _, outputId := range e.OutputIds {
			s.outputs[outputId].state = OutputStateActive
			s.outputs[outputId].sessId = e.SessionId
		}

		// TODO: mutate outputs state
		//fmt.Println("=== UNHANDLED EVENT from sink", e)
	default:
		fmt.Printf("@@@ 1 unknown event %#v %s\n", e, reflect.TypeOf(e))
	}
}
