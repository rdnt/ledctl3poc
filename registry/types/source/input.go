package source

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"
)

type Input struct {
	id   uuid.UUID
	name string

	state  InputState
	sessId uuid.UUID

	sinks  []sinkConfig
	schema map[string]any
	cfg    map[string]any
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

func NewInput(id uuid.UUID, name string, schema map[string]any) *Input {
	return &Input{
		id:     id,
		name:   name,
		state:  InputStateIdle,
		schema: schema,
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

func (i *Input) ApplyConfig(cfg map[string]any) error {
	schema := gojsonschema.NewGoLoader(i.schema)
	document := gojsonschema.NewGoLoader(cfg)

	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return err
	}

	if !result.Valid() {
		fmt.Printf("Invalid input config:\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return errors.New("invalid input config")
	}

	i.cfg = cfg
	return nil
}
