package event

import (
	"image/color"

	"ledctl3/pkg/uuid"
)

type Data struct {
	SessionId uuid.UUID
	Outputs   []DataOutput
}

type DataOutput struct {
	Id  uuid.UUID
	Pix []color.Color
}
