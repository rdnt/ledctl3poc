package event

import "ledctl3/pkg/uuid"

type OutputDisconnected struct {
	Id uuid.UUID
}
