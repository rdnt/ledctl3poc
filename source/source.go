package source

import (
	"fmt"
	"github.com/google/uuid"
	"ledctl3/source/types"
	"ledctl3/source/types/event"
	"net"
	"sync"
)

type Source struct {
	mux        sync.Mutex
	id         uuid.UUID
	address    net.Addr
	state      types.State
	sessionId  string
	leds       int
	visualizer event.Visualizer
}

func (s *Source) ProcessEvent(e event.Event) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch e := e.(type) {
	case event.SetActiveEvent:
		s.handleSetActiveEvent(e)
	case event.SetIdleEvent:
		s.handleSetIdleEvent(e)
	default:
		fmt.Println("unknown event", e)
	}
}

func (s *Source) handleSetActiveEvent(e event.SetActiveEvent) {
	if s.state == types.StateIdle {
		s.sessionId = e.SessionId

		if s.leds != e.Leds {
			fmt.Println("engine restart")
			s.leds = e.Leds
		}

		if s.visualizer != e.Visualizer {
			fmt.Println("set visualizer", e.Visualizer)
			s.visualizer = e.Visualizer
		}

		s.state = types.StateActive
	}
}

func (s *Source) handleSetIdleEvent(_ event.SetIdleEvent) {
	if s.state == types.StateActive {
		s.sessionId = ""
		s.state = types.StateIdle

		fmt.Println("===", s.visualizer)
		if s.visualizer != "" {
			fmt.Println("stop visualizer")
			s.visualizer = event.VisualizerNone
		}
	}
}

func New(address net.Addr) *Source {
	return &Source{
		id:         uuid.New(),
		address:    address,
		state:      types.StateIdle,
		visualizer: event.VisualizerNone,
	}
}
