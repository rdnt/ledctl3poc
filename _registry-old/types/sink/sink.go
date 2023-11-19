package sink

import (
	"fmt"
	"net"
	"reflect"

	"ledctl3/pkg/uuid"

	"ledctl3/event"
	regevent "ledctl3/event"
)

type Sink struct {
	Id      uuid.UUID
	Addr    net.Addr
	Name    string
	Outputs map[uuid.UUID]*Output
}

func New(id uuid.UUID, addr net.Addr) *Sink {
	return &Sink{
		Id:   id,
		Addr: addr,
	}
}

func (s *Sink) Leds() int {
	var leds int
	for _, dev := range s.Outputs {
		leds += dev.Leds
	}

	return leds
}

type Calibration struct {
	R float64
	G float64
	B float64
	A float64
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
		"sink{Id: %s, Name: %s, Leds: %Id, Calibration: %v}",
		s.Id, s.Name, s.Leds(), s.Calibration(),
	)
}

func (s *Sink) Process(e event.EventIface) {
	switch e := e.(type) {
	case regevent.SetSinkActive:
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
