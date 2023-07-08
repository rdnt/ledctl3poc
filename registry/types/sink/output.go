package sink

import (
	"fmt"

	"github.com/google/uuid"

	"ledctl3/registry"
)

type Output struct {
	id   uuid.UUID
	name string

	state  registry.OutputState
	sessId uuid.UUID

	leds        int
	calibration map[int]registry.Calibration
}

func NewOutput(id uuid.UUID, name string) *Output {
	return &Output{
		id:    id,
		name:  name,
		state: registry.OutputStateIdle,
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

func (o *Output) Calibration() map[int]registry.Calibration {
	return o.calibration
}

func (o *Output) State() registry.OutputState {
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
