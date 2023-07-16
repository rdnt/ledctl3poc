package video

import (
	"fmt"

	"github.com/google/uuid"

	"ledctl3/source"
)

type ScreenCapture struct {
	id     uuid.UUID
	events chan source.UpdateEvent
}

func (s *ScreenCapture) AssistedSetup() (map[string]any, error) {
	return nil, nil
}

func (s *ScreenCapture) Id() uuid.UUID {
	return s.id
}

func (s *ScreenCapture) Start(cfg source.SinkConfig) error {
	fmt.Printf("## starting video source with config: %#v\n", cfg)
	return nil
}

func (s *ScreenCapture) Events() chan source.UpdateEvent {
	return s.events
}

func (s *ScreenCapture) Stop() error {
	return nil
}

func New() *ScreenCapture {
	return &ScreenCapture{
		id:     uuid.New(),
		events: make(chan source.UpdateEvent),
	}
}
