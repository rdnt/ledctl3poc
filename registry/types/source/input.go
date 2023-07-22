package source

import (
	"errors"
	"fmt"

	"ledctl3/pkg/uuid"
	"github.com/xeipuuv/gojsonschema"
)

type Input struct {
	Id   uuid.UUID
	Name string

	State     InputState
	SessionId uuid.UUID

	Sinks   []SinkConfig
	Schema  map[string]any
	Config  InputConfig
	Configs map[uuid.UUID]InputConfig
}

type InputConfig struct {
	Id   uuid.UUID
	Name string
	Cfg  map[string]any
}

type InputState string

const (
	InputStateIdle   InputState = "idle"
	InputStateActive InputState = "active"
)

type SinkConfig struct {
	Id      uuid.UUID
	Outputs []OutputConfig
}

type OutputConfig struct {
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
		Id:      id,
		Name:    name,
		State:   InputStateIdle,
		Schema:  schema,
		Configs: map[uuid.UUID]InputConfig{},
	}
}

func (i *Input) String() string {
	return fmt.Sprintf(
		"input{OutputId: %s, Name: %s, State: %s}",
		i.Id, i.Name, i.State,
	)
}

func (i *Input) AddConfig(name string, cfg map[string]any) (InputConfig, error) {
	schema := gojsonschema.NewGoLoader(i.Schema)
	document := gojsonschema.NewGoLoader(cfg)

	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return InputConfig{}, err
	}

	if !result.Valid() {
		fmt.Printf("Invalid input Config:\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return InputConfig{}, errors.New("invalid input Config")
	}

	if name == "" {
		name = "New Config" // TODO: incremental number suffix
	}

	conf := InputConfig{
		Id:   uuid.New(),
		Name: name,
		Cfg:  cfg,
	}

	i.Configs[conf.Id] = conf

	return conf, nil
}

func (i *Input) ApplyConfig(id uuid.UUID) error {
	cfg, ok := i.Configs[id]
	if !ok {
		return errors.New("invalid config OutputId")
	}

	i.Config = cfg

	return nil
}
