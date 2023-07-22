package uuid

import (
	"github.com/google/uuid"
)

type UUID string

func New() UUID {
	return UUID(uuid.New().String())
}

func (u UUID) String() string {
	return string(u)
}

var Nil = UUID(uuid.Nil.String())

func Parse(s string) (UUID, error) {
	uid, err := uuid.Parse(s)
	if err != nil {
		return Nil, err
	}

	return UUID(uid.String()), nil
}
