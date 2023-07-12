package source

import (
	"fmt"

	"github.com/google/uuid"
)

type Input struct {
	id   uuid.UUID
	name string

	state  InputState
	sessId uuid.UUID

	sinks []sinkConfig
}

type InputState string

const (
	InputStateIdle   InputState = "idle"
	InputStateActive InputState = "active"
)

type sinkConfig struct {
	Id      uuid.UUID
	Outputs []outputConfig
}

type outputConfig struct {
	Id   uuid.UUID
	Leds int
}

type sink struct {
	id      uuid.UUID
	outputs []output
}

type output struct {
	id   uuid.UUID
	leds int
}

func NewInput(id uuid.UUID, name string) *Input {
	return &Input{
		id:    id,
		name:  name,
		state: InputStateIdle,
	}
}

func (i *Input) Id() uuid.UUID {
	return i.id
}

func (i *Input) Name() string {
	return i.name
}

func (i *Input) State() InputState {
	return i.state
}

func (i *Input) SessionId() uuid.UUID {
	return i.sessId
}

func (i *Input) String() string {
	return fmt.Sprintf(
		"input{id: %s, name: %s, state: %s}",
		i.id, i.name, i.state,
	)
}
