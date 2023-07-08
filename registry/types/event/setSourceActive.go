package event

import (
	"github.com/google/uuid"

	"ledctl3/pkg/event"
)

// SetSourceActiveEvent instructs a source to become active for the specified inputs (keys on the sinks map)
// towards the outputs in an array of sinks for each input (e.g. 1 input to N outputs that reside in
// different sinks)
type SetSourceActiveEvent struct {
	Event
	SessionId uuid.UUID                                `json:"sessionId"`
	Sinks     map[uuid.UUID][]SetSourceActiveEventSink `json:"sinks"`
}

type SetSourceActiveEventSink struct {
	Id        uuid.UUID   `json:"id"`
	Address   string      `json:"address"`
	OutputIds []uuid.UUID `json:"outputIds"`
}

func (e SetSourceActiveEvent) Type() event.Type {
	return SetSourceActive
}
