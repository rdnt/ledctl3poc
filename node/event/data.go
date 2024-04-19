package event

import (
	"image/color"
	"time"

	"github.com/google/uuid"
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
