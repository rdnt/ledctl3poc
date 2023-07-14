package video

import (
	"fmt"

	"github.com/google/uuid"

	"ledctl3/source"
)

type VideoSource struct {
	id     uuid.UUID
	events chan source.UpdateEvent
}

func (v *VideoSource) Id() uuid.UUID {
	return v.id
}

func (v *VideoSource) Start(cfg source.SinkConfig) error {
	fmt.Printf("## starting video source with config: %#v\n", cfg)
	return nil
}

func (v *VideoSource) Events() chan source.UpdateEvent {
	return v.events
}

func (v *VideoSource) Stop() error {
	return nil
}

func New() *VideoSource {
	return &VideoSource{
		id:     uuid.New(),
		events: make(chan source.UpdateEvent),
	}
}
