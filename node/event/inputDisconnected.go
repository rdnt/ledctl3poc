package event

import "ledctl3/pkg/uuid"

type InputDisconnected struct {
	Id uuid.UUID
}
