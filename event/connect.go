package event

import "ledctl3/pkg/uuid"

type ConnectEvent struct {
	Event
	Id uuid.UUID `json:"id"`
}

func (e ConnectEvent) Type() Type {
	return Connect
}
