package event

import "ledctl3/pkg/uuid"

type DriverConfig struct {
	DriverId uuid.UUID
	Config   []byte
	Schema   []byte
}
