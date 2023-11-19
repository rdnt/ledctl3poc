package _sink_old

import (
	"fmt"
	"image/color"
	"net"
	"sync"

	gcolor "github.com/gookit/color"
	"github.com/samber/lo"

	"ledctl3/pkg/uuid"

	"ledctl3/event"

	"ledctl3/_source-old/types"
)

type Output interface {
	Id() uuid.UUID
	Start() error
	//Render(UpdateEvent)
	Stop() error
	Leds() int
}

type UpdateEvent struct {
	Pix []color.Color
}

type Message struct {
	Addr  net.Addr
	Event event.EventIface
}

type Sink struct {
	mux sync.Mutex
	id  uuid.UUID
	//address    net.Addr
	state     types.State
	sessionId uuid.UUID
	outputs   map[uuid.UUID]Output
	events    chan Message
}

func New(id uuid.UUID) *Sink {
	s := &Sink{
		id: id,
		//address: address,
		state:   types.StateIdle,
		outputs: make(map[uuid.UUID]Output),
		events:  make(chan Message),
	}

	return s
}

func (s *Sink) Id() uuid.UUID {
	return s.id
}

func (s *Sink) ProcessEvent(addr net.Addr, e event.EventIface) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch e := e.(type) {
	case event.ListCapabilities:
		fmt.Printf("%s -> sink: recv ListCapabilities\n", addr)
		s.handleListCapabilitiesEvent(addr, e)
	case event.SetSinkActive:
		fmt.Printf("%s -> sink: recv SetSinkActive\n", addr)
		s.handleSetActiveEvent(addr, e)
	case event.Data:
		fmt.Printf("%s -> sink: recv Data\n", addr)
		s.handleDataEvent(addr, e)
	default:
		fmt.Println("unknown event", e)
	}
}

func (s *Sink) handleSetActiveEvent(addr net.Addr, e event.SetSinkActive) {
	if len(e.OutputIds) == 0 {
		return
	}

	//fmt.Println("=== sink activating")
}

func (s *Sink) handleDataEvent(addr net.Addr, e event.Data) {
	for _, output := range e.Outputs {
		out := "\n"
		for _, c := range output.Pix {
			r, g, b, _ := c.RGBA()
			//if r != 0 {
			//	g = 0
			//	b = 0
			//}
			out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
		}
		// @@@ DEBUG
		fmt.Print(out)
	}
}

func (s *Sink) handleListCapabilitiesEvent(addr net.Addr, _ event.ListCapabilities) {
	s.events <- Message{
		Addr: addr,
		Event: event.Capabilities{
			Event:  event.Event{Type: event.Capabilities},
			Id:     s.id,
			Inputs: []event.CapabilitiesInput{},
			Outputs: lo.Map(lo.Values(s.outputs), func(output Output, _ int) event.CapabilitiesOutput {
				return event.CapabilitiesOutput{
					Id:   output.Id(),
					Leds: output.Leds(),
				}
			}),
		},
	}
}

func (s *Sink) AddOutput(o Output) {
	s.outputs[o.Id()] = o

	//go func() {
	//	// forward events from input to the network~
	//	for e := range o.Messages() {
	//		var outputs []event.DataOutput
	//		for _, output := range e.Outputs {
	//			outputs = append(outputs, event.DataOutput{
	//				Id:  output.Id,
	//				Pix: output.Pix,
	//			})
	//		}
	//
	//		s.events <- event.Data{
	//			Payload:     event.Payload{Type: event.Data, Addr: e.Id},
	//			SessionId: s.sessionId,
	//			Outputs:   outputs,
	//		}
	//	}
	//}()
}

func (s *Sink) Messages() <-chan Message {
	return s.events
}
