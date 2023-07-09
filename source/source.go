package source

import (
	"fmt"
	"image/color"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"ledctl3/event"

	"ledctl3/source/types"
)

type Input interface {
	Id() string
	Start() error
	Events() chan UpdateEvent
	Stop() error
}

type UpdateEvent struct {
	Pix     []color.Color
	Latency time.Duration
}

type Source struct {
	mux       sync.Mutex
	id        uuid.UUID
	address   net.Addr
	state     types.State
	sessionId uuid.UUID
	inputs    map[uuid.UUID]Input
	events    chan event.EventIface
	inputCfgs map[uuid.UUID][]sinkConfig
}

type sinkConfig struct {
	Id      uuid.UUID
	Outputs []outputConfig
}

type outputConfig struct {
	Id   uuid.UUID
	Leds int
}

func New(address net.Addr, inputs map[uuid.UUID]Input) *Source {
	s := &Source{
		id:        uuid.New(),
		address:   address,
		state:     types.StateIdle,
		inputs:    inputs,
		events:    make(chan event.EventIface),
		inputCfgs: map[uuid.UUID][]sinkConfig{},
	}

	return s
}

func (s *Source) Events() <-chan event.EventIface {
	return s.events
}

func (s *Source) ProcessEvent(e event.EventIface) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch e := e.(type) {
	case event.SetSourceActiveEvent:
		fmt.Printf("-> source %s: recv SetSourceActiveEvent\n", s.id)
		s.handleSetActiveEvent(e)
	case event.SetSourceIdleEvent:
		fmt.Printf("-> source %s: recv SetSourceIdleEvent\n", s.id)
		s.handleSetIdleEvent(e)
	default:
		fmt.Println("unknown event", e)
	}
}

func (s *Source) handleSetActiveEvent(e event.SetSourceActiveEvent) {
	if s.state == types.StateIdle {
		//fmt.Println("Initializing session", e.SessionId, e.Sinks)

		s.state = types.StateActive
		s.sessionId = e.SessionId

		for _, input := range e.Inputs {
			s.inputCfgs[input.Id] = lo.Map(input.Sinks, func(sink event.SetSourceActiveEventSink, index int) sinkConfig {
				return sinkConfig{
					Id: sink.Id,
					Outputs: lo.Map(sink.Outputs, func(output event.SetSourceActiveEventOutput, _ int) outputConfig {
						return outputConfig(output)
					}),
				}
			})
		}

		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")

		// TODO: validate there are no sourceIds present in the event but not present on the driver

		for _, input := range e.Inputs {
			input := input
			go func() {
				for {
					for _, sink := range s.inputCfgs[input.Id] {
						var outputs []event.DataEventOutput
						for _, output := range sink.Outputs {
							pix := make([]byte, output.Leds*4)

							outputs = append(outputs, event.DataEventOutput{
								Id:  output.Id,
								Pix: pix,
							})
						}

						s.events <- event.DataEvent{
							Event:     event.Event{Type: event.Data, DevId: sink.Id},
							SessionId: s.sessionId,
							Outputs:   outputs,
						}
					}

					time.Sleep(1 * time.Second)
				}
			}()
		}

		//for _, input := range e.Inputs {
		//	//fmt.Println("starting input", inputId, "with sink cfgs", sinkCfgs)
		//	//fmt.Printf("-> source %s: start input %s\n", s.id, inputId)
		//	_ = inputId
		//	_ = sinkCfgs
		//
		//	// TODO: inputs is never populated right now. when to do that? on event start? on startup?
		//	// persist it?
		//	//err := s.inputs[inputId].Start()
		//	//if err != nil {
		//	//	fmt.Println("failed to start source", err)
		//	//}
		//
		//	//go func() {
		//	//	for _, input := range s.inputs {
		//	//		for ue := range input.Events() {
		//	//			for _, sinkCfg := range sinkCfgs {
		//	//				e := regevent.SetLedsEvent{
		//	//					SessionId: uuid.Nil, // TODO
		//	//					SinkId:    sinkCfg.Id,
		//	//					// TODO: apply calibration to pix and pass clamped as byte array
		//	//					Pix: nil, // calibrate(ue.Pix)
		//	//				}
		//	//
		//	//				_ = ue
		//	//
		//	//				s.events <- e
		//	//			}
		//	//
		//	//		}
		//	//	}
		//	//}()
		//
		//}
	}
}

func (s *Source) handleSetIdleEvent(_ event.SetSourceIdleEvent) {
	if s.state == types.StateActive {
		s.state = types.StateIdle
		s.sessionId = uuid.Nil

		//for _, src := range s.inputs {
		//	err := src.Stop()
		//	if err != nil {
		//		fmt.Println("failed to stop source", err)
		//	}
		//}
	}
}

func (s *Source) Id() uuid.UUID {
	return s.id
}
