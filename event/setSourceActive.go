package event

import (
	"github.com/google/uuid"
)

// SetSourceActiveEvent instructs a source to become active for the specified inputs (keys on the sinks map)
// towards the outputs in an array of sinks for each input (e.g. 1 input to N outputs that reside in
// different sinks)
type SetSourceActiveEvent struct {
	Event
	SessionId uuid.UUID                   `json:"sessionId"`
	Inputs    []SetSourceActiveEventInput `json:"inpus"`
}

type SetSourceActiveEventInput struct {
	Id    uuid.UUID                  `json:"id"`
	Sinks []SetSourceActiveEventSink `json:"sinks"`
}

type SetSourceActiveEventSink struct {
	Id      uuid.UUID                    `json:"id"`
	Outputs []SetSourceActiveEventOutput `json:"outputs"`
}

type SetSourceActiveEventOutput struct {
	Id     uuid.UUID      `json:"id"`
	Config map[string]any `json:"config"`
	Leds   int            `json:"leds"`
}

func (e SetSourceActiveEvent) Type() Type {
	return SetSourceActive
}
