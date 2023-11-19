package event

import "ledctl3/pkg/uuid"

type SetInputConfig struct {
	InputId uuid.UUID
	Config  any
}
