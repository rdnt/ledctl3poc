package event

import (
	"ledctl3/pkg/uuid"
)

type SetSourceIdle struct {
	Inputs []SetSourceIdleInput
}

type SetSourceIdleInput struct {
	InputId   uuid.UUID
	OutputIds []uuid.UUID
}
