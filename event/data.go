package event

import (
	"github.com/google/uuid"
)

type DataEvent struct {
	Event
	SessionId uuid.UUID         `json:"sessionId"`
	Outputs   []DataEventOutput `json:"outputs"`
}

type DataEventOutput struct {
	Id  uuid.UUID `json:"id"`
	Pix []byte    `json:"pix"`
}

func (e DataEvent) Type() Type {
	return Data
}
