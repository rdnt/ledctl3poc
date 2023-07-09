package event

import "github.com/google/uuid"

type ConnectEvent struct {
	Event
	Id uuid.UUID `json:"id"`
}

func (e ConnectEvent) Type() Type {
	return Connect
}
