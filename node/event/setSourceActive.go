package event

import "github.com/google/uuid"

type SetSourceActive struct {
	Inputs []SetSourceActiveInput
}

type SetSourceActiveInput struct {
	Id      uuid.UUID
	Outputs []SetSourceActiveOutput
}

type SetSourceActiveOutput struct {
	Id     uuid.UUID
	SinkId uuid.UUID
	Leds   int
	Config map[string]any
}

func (e SetSourceActive) Type() string {
	return TypeSetSourceActive
}
