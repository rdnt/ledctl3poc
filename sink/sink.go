package sink

import (
	"fmt"
	"image/color"
	"sync"

	"github.com/google/uuid"
	gcolor "github.com/gookit/color"
	"github.com/samber/lo"

	"ledctl3/event"

	"ledctl3/source/types"
)

type Output interface {
	Id() uuid.UUID
	Start() error
	//Handle(UpdateEvent)
	Stop() error
	Leds() int
}

type UpdateEvent struct {
	Pix []color.Color
}

type Sink struct {
	mux sync.Mutex
	id  uuid.UUID
	//address    net.Addr
	state      types.State
	sessionId  uuid.UUID
	outputs    map[uuid.UUID]Output
	events     chan event.EventIface
	registryId uuid.UUID
}

func New(registryId uuid.UUID) *Sink {
	s := &Sink{
		id: uuid.New(),
		//address: address,
		state:      types.StateIdle,
		outputs:    make(map[uuid.UUID]Output),
		events:     make(chan event.EventIface),
		registryId: registryId,
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
	case event.ListCapabilitiesEvent:
		fmt.Printf("-> sink %s: recv ListCapabilitiesEvent\n", s.id)
		s.handleListCapabilitiesEvent(e)
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

func (s *Sink) Connect() {
	s.events <- event.ConnectEvent{
		Event: event.Event{Type: event.Connect, DevId: s.registryId},
		Id:    s.id,
	}
}

func (s *Sink) handleListCapabilitiesEvent(_ event.ListCapabilitiesEvent) {
	s.events <- event.CapabilitiesEvent{
		Event:  event.Event{Type: event.Capabilities, DevId: s.registryId},
		Id:     s.id,
		Inputs: []event.CapabilitiesEventInput{},
		Outputs: lo.Map(lo.Values(s.outputs), func(output Output, _ int) event.CapabilitiesEventOutput {
			return event.CapabilitiesEventOutput{
				Id:   output.Id(),
				Leds: output.Leds(),
			}
		}),
	}
}

func (s *Sink) AddOutput(o Output) {
	s.outputs[o.Id()] = o

	//go func() {
	//	// forward events from input to the network~
	//	for e := range o.Events() {
	//		var outputs []event.DataEventOutput
	//		for _, output := range e.Outputs {
	//			outputs = append(outputs, event.DataEventOutput{
	//				Id:  output.Id,
	//				Pix: output.Pix,
	//			})
	//		}
	//
	//		s.events <- event.DataEvent{
	//			Event:     event.Event{Type: event.Data, DevId: e.SinkId},
	//			SessionId: s.sessionId,
	//			Outputs:   outputs,
	//		}
	//	}
	//}()
}

func (s *Sink) Events() <-chan event.EventIface {
	return s.events
}
