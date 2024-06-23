package types

import (
	"image/color"
	"time"

	"github.com/google/uuid"
)

type UpdateEvent struct {
	NodeId  uuid.UUID
	Outputs []UpdateEventOutput
	Latency time.Duration
}

type UpdateEventOutput struct {
	SinkId   uuid.UUID
	OutputId uuid.UUID
	Pix      []color.NRGBA
}
