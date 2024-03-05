package event

import "ledctl3/pkg/uuid"

type OutputDisconnected struct {
	Id uuid.UUID
}

func (e OutputDisconnected) Type() string {
	return TypeOutputDisconnected
}
