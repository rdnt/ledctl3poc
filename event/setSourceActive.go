package event

import (
	"ledctl3/pkg/uuid"
)

type SetSourceActive struct {
	Inputs []SetSourceActiveInput
}

type SetSourceActiveInput struct {
	Id    uuid.UUID
	Sinks []SetSourceActiveSink
}

type SetSourceActiveSink struct {
	Id      uuid.UUID
	Outputs []SetSourceActiveOutput
}

type SetSourceActiveOutput struct {
	Id     uuid.UUID
	Config map[string]any
	Leds   int
}
