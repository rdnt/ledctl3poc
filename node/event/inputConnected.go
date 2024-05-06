package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type InputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Schema   json.RawMessage
	Config   json.RawMessage
}

func (e InputConnected) Type() string {
	return TypeInputConnected
}
