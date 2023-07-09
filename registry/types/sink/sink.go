package sink

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/event"
	regevent "ledctl3/event"
	"ledctl3/registry"
)

type Sink struct {
	id   uuid.UUID
	name string

	outputs map[uuid.UUID]*Output

	send func(event.EventIface) error
	recv func() <-chan event.EventIface
}

func NewSink(id uuid.UUID, name string, outputs map[uuid.UUID]*Output, send func(event.EventIface) error, recv func() <-chan event.EventIface) *Sink {
	return &Sink{
		id:      id,
		name:    name,
		outputs: outputs,
		send:    send,
		recv:    recv,
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

func (s *Sink) Calibration() map[int]registry.Calibration {
	calib := make(map[int]registry.Calibration)

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

func (s *Sink) Outputs() map[uuid.UUID]registry.Output {
	outputs := make(map[uuid.UUID]registry.Output)

	for id, output := range s.outputs {
		outputs[id] = output
	}

	return outputs
}

func (s *Sink) Handle(e event.EventIface) error {
	err := s.send(e)
	if err != nil {
		return err
	}

	s.processEvent(e)
	return nil
}

func (s *Sink) processEvent(e event.EventIface) {
	switch e := e.(type) {
	case regevent.SetSinkActiveEvent:
		for _, outputId := range e.OutputIds {
			s.outputs[outputId].state = registry.OutputStateActive
			s.outputs[outputId].sessId = e.SessionId
		}

		// TODO: mutate outputs state
		//fmt.Println("=== UNHANDLED EVENT from sink", e)
	default:
		fmt.Printf("@@@ 1 unknown event %#v %s\n", e, reflect.TypeOf(e))
	}
}

func (s *Sink) Events() <-chan event.EventIface {
	return s.recv()
}
