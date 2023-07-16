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
	Id   uuid.UUID
	Name string

	Outputs map[uuid.UUID]*Output
}

func NewSink(id uuid.UUID, name string, outputs map[uuid.UUID]*Output) *Sink {
	return &Sink{
		Id:      id,
		Name:    name,
		Outputs: outputs,
	}
}

func (s *Sink) Leds() int {
	var leds int
	for _, dev := range s.Outputs {
		leds += dev.Leds
	}

	return leds
}

func (s *Sink) Calibration() map[int]Calibration {
	calib := make(map[int]Calibration)

	var acc int
	for _, out := range s.Outputs {
		for i, c := range out.Calibration {
			calib[i+acc] = c
		}

		acc += out.Leds
	}

	return calib
}

func (s *Sink) String() string {
	return fmt.Sprintf(
		"sink{OutputId: %s, Name: %s, Leds: %OutputId, Calibration: %v}",
		s.Id, s.Name, s.Leds(), s.Calibration(),
	)
}

func (s *Sink) Process(e event.EventIface) {
	switch e := e.(type) {
	case regevent.SetSinkActiveEvent:
		for _, outputId := range e.OutputIds {
			s.Outputs[outputId].State = OutputStateActive
			s.Outputs[outputId].SessionId = e.SessionId
		}

		// TODO: mutate Outputs State
		//fmt.Println("=== UNHANDLED EVENT from sink", e)
	default:
		fmt.Printf("@@@ 1 unknown event %#v %s\n", e, reflect.TypeOf(e))
	}
}
