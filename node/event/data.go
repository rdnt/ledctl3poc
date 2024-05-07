package event

import (
	"image/color"
	"time"

	"github.com/google/uuid"
)

type Data struct {
	NodeId  uuid.UUID
	SinkId  uuid.UUID
	Outputs []DataOutput
	Latency time.Duration
}

type DataOutput struct {
	OutputId uuid.UUID
	Pix      []color.NRGBA
}

func (e Data) Type() string {
	return TypeData
}
