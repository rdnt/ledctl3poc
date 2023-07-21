package types

import (
	"image/color"
	"time"

	"github.com/google/uuid"
)

type UpdateEvent struct {
	SinkId  uuid.UUID
	Outputs []UpdateEventOutput
	Latency time.Duration
}

type UpdateEventOutput struct {
	OutputId uuid.UUID
	Pix      []color.Color
}
