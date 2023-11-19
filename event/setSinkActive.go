package event

import (
	"ledctl3/pkg/uuid"
)

type SetSinkActive struct {
	SessionId uuid.UUID
	OutputIds []uuid.UUID
}
