package event

import "ledctl3/pkg/uuid"

type Connect struct {
	Id      uuid.UUID
	Drivers []ConnectDriver
}

type ConnectDriver struct {
	Id     uuid.UUID
	Config []byte
	Schema []byte
}
