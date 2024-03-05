package event

import (
	"image/color"
	"time"

	"ledctl3/pkg/uuid"
)

type Data struct {
	SinkId  uuid.UUID
	Outputs []DataOutput
	Latency time.Duration
}

type DataOutput struct {
	OutputId uuid.UUID
	Pix      []color.Color
}

func (e Data) Type() string {
	return TypeData
}
