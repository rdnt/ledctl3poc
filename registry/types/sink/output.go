package sink

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Output struct {
	Id   uuid.UUID
	Name string

	State     OutputState
	SessionId uuid.UUID

	Leds        int
	Calibration map[int]Calibration
}

type OutputState string

const (
	OutputStateIdle   OutputState = "idle"
	OutputStateActive OutputState = "active"
)

func NewOutput(id uuid.UUID, name string, leds int) *Output {
	return &Output{
		Id:    id,
		Name:  name,
		State: OutputStateIdle,
		Leds:  leds,
	}
}

func (o *Output) String() string {
	return fmt.Sprintf(
		"output{OutputId: %s, Name: %s, Leds: %OutputId, Calibration: %v, State: %s}",
		o.Id, o.Name, o.Leds, o.Calibration, o.State,
	)
}
