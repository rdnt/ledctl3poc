package types

import (
	"image/color"
	"time"

	"ledctl3/pkg/uuid"
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
