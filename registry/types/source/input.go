package source

import (
	"fmt"

	"github.com/google/uuid"

	"ledctl3/registry"
)

type Input struct {
	id   uuid.UUID
	name string

	state  registry.InputState
	sessId uuid.UUID

	sinks []sinkConfig
}

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
		state: registry.InputStateIdle,
	}
}

func (i *Input) Id() uuid.UUID {
	return i.id
}

func (i *Input) Name() string {
	return i.name
}

func (i *Input) State() registry.InputState {
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

func (i *Input) Start() {
	fmt.Println("starting input with cfg", i.sinks)
	//i.state = registry.InputStateActive
}
