package event

import (
	"image/color"

	"github.com/google/uuid"
)

type DataEvent struct {
	Event
	SessionId uuid.UUID         `json:"sessionId"`
	Outputs   []DataEventOutput `json:"outputs"`
}

type DataEventOutput struct {
	Id  uuid.UUID     `json:"id"`
	Pix []color.Color `json:"pix"`
}

func (e DataEvent) Type() Type {
	return Data
}
