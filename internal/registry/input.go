package registry

import (
	"encoding/json"
	"fmt"

	"ledctl3/pkg/uuid"
)

type Input struct {
	Id        uuid.UUID       `json:"id"`
	DriverId  uuid.UUID       `json:"driverId"`
	Schema    json.RawMessage `json:"schema"`
	Config    json.RawMessage `json:"config"`
	Connected bool            `json:"connected"`
}

func NewInput(id, driverId uuid.UUID, schema, config []byte, connected bool) *Input {
	return &Input{
		Id:        id,
		DriverId:  driverId,
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
