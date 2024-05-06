package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type OutputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Leds     int
	Schema   json.RawMessage
	Config   json.RawMessage
}

func (e OutputConnected) Type() string {
	return TypeOutputConnected
}
