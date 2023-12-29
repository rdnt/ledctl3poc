package event

import "ledctl3/pkg/uuid"

type SetDriverConfig struct {
	DriverId uuid.UUID
	Config   []byte
}
