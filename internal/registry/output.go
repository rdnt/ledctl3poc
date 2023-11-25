package registry

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Output struct {
	Id        uuid.UUID      `json:"id"`
	Leds      int            `json:"leds"`
	Schema    map[string]any `json:"schema"`
	Config    map[string]any `json:"config"`
	Connected bool           `json:"-"`
}

func NewOutput(id uuid.UUID, leds int, schema, config map[string]any, connected bool) *Output {
	return &Output{
		Id:        id,
		Leds:      leds,
		Schema:    schema,
		Config:    config,
		Connected: connected,
	}
}

func (out *Output) Connect() {
	fmt.Println("output Connected:", out.Id)

	out.Connected = true
}

func (out *Output) Disconnect() {
	fmt.Println("output disconnected:", out.Id)

	out.Connected = false
}
