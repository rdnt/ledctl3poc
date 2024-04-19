package event

import "github.com/google/uuid"

type Authenticate struct {
	Id uuid.UUID
}

func (e Authenticate) Type() string {
	return TypeAuthenticate
}
