package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type InputConnected struct {
	Id       uuid.UUID
	SourceId uuid.UUID
	Schema   json.RawMessage
}

func (e InputConnected) Type() string {
	return TypeInputConnected
}
