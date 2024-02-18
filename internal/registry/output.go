package registry

import (
	"encoding/json"
	"fmt"

	"ledctl3/pkg/uuid"
)

type Output struct {
	Id        uuid.UUID       `json:"id"`
	DriverId  uuid.UUID       `json:"driverId"`
	Leds      int             `json:"leds"`
	Schema    json.RawMessage `json:"schema"`
	Config    json.RawMessage `json:"config"`
	Connected bool            `json:"connected"`
}

func NewOutput(id, driverId uuid.UUID, leds int, schema, config []byte, connected bool) *Output {
	return &Output{
		Id:        id,
		DriverId:  driverId,
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
