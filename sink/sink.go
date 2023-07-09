package sink

import (
	"fmt"
	"image/color"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	gcolor "github.com/gookit/color"

	"ledctl3/event"

	"ledctl3/source/types"
)

type Output interface {
	Id() string
	Start() error
	Handle(UpdateEvent)
	Stop() error
}

type UpdateEvent struct {
	Pix     []color.Color
	Latency time.Duration
}

type Sink struct {
	mux       sync.Mutex
	id        uuid.UUID
	address   net.Addr
	state     types.State
	sessionId uuid.UUID
	outputs   map[uuid.UUID]Output
	events    chan event.EventIface
}

func New(address net.Addr) *Sink {
	s := &Sink{
		id:      uuid.New(),
		address: address,
		state:   types.StateIdle,
		outputs: make(map[uuid.UUID]Output),
		events:  make(chan event.EventIface),
	}

	return s
}

func (s *Sink) Id() uuid.UUID {
	return s.id
}

func (s *Sink) ProcessEvent(e event.EventIface) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch e := e.(type) {
	case event.SetSinkActiveEvent:
		fmt.Printf("-> sink %s: recv SetSinkActiveEvent\n", s.id)
		s.handleSetActiveEvent(e)
	case event.DataEvent:
		fmt.Printf("-> sink %s: recv DataEvent\n", s.id)
		s.handleDataEvent(e)
	default:
		fmt.Println("unknown event", e)
	}
}

func (s *Sink) handleSetActiveEvent(e event.SetSinkActiveEvent) {
	if len(e.OutputIds) == 0 {
		return
	}

	//fmt.Println("=== sink activating")
}

func (s *Sink) handleDataEvent(e event.DataEvent) {
	//fmt.Println("SINK: HANDLING DATA EVENT", e.Outputs)
	for _, output := range e.Outputs {
		out := ""
		for _, c := range output.Pix {
			r, g, b, _ := c.RGBA()
			if r != 0 {
				g = 0
				b = 0
			}
			out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
		}
		fmt.Println(out)
	}
}
