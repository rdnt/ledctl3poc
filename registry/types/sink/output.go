package sink

import (
	"fmt"

	"github.com/google/uuid"
)

type Output struct {
	id   uuid.UUID
	name string

	state  OutputState
	sessId uuid.UUID

	leds        int
	calibration map[int]Calibration
}

type OutputState string

const (
	OutputStateIdle   OutputState = "idle"
	OutputStateActive OutputState = "active"
)

func NewOutput(id uuid.UUID, name string, leds int) *Output {
	return &Output{
		id:    id,
		name:  name,
		state: OutputStateIdle,
		leds:  leds,
	}
}

func (o *Output) Id() uuid.UUID {
	return o.id
}

func (o *Output) Name() string {
	return o.name
}

func (o *Output) Leds() int {
	return o.leds
}

func (o *Output) Calibration() map[int]Calibration {
	return o.calibration
}

func (o *Output) State() OutputState {
	return o.state
}

func (o *Output) SessionId() uuid.UUID {
	return o.sessId
}

func (o *Output) String() string {
	return fmt.Sprintf(
		"output{id: %s, name: %s, leds: %d, calibration: %v, state: %s}",
		o.id, o.name, o.Leds(), o.Calibration(), o.state,
	)
}
