package registry

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Input struct {
	Id        uuid.UUID      `json:"id"`
	Schema    map[string]any `json:"schema"`
	Config    map[string]any `json:"config"`
	Connected bool           `json:"-"`
}

func NewInput(id uuid.UUID, schema, config map[string]any, connected bool) *Input {
	return &Input{
		Id:        id,
		Schema:    schema,
		Config:    config,
		Connected: connected,
	}
}

func (in *Input) Connect() {
	fmt.Println("input Connected:", in.Id)

	in.Connected = true
}

func (in *Input) Disconnect() {
	fmt.Println("input disconnected:", in.Id)

	in.Connected = false
}
