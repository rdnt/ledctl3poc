package registry

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Input struct {
	Id        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Connected bool      `json:"-"`
}

func NewInput(id uuid.UUID, typ string, connected bool) *Input {
	return &Input{
		Id:        id,
		Type:      typ,
		Connected: connected,
	}
}

func (in *Input) Connect() {
	fmt.Println("input Connected:", in.Id)

	in.Connected = true
}

func (in *Input) Disconnect() {
	fmt.Println("input disconnected:", in.Id)

	in.Connected = false
}
