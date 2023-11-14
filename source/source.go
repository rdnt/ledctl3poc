package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/samber/lo"

	"ledctl3/pkg/uuid"

	"ledctl3/event"
	"ledctl3/source/types"
)

type Input interface {
	Id() uuid.UUID
	Start(cfg types.InputConfig) error
	Events() <-chan types.UpdateEvent
	Stop() error
	Schema() map[string]any
	//ApplyConfig(cfg map[string]any) error
	AssistedSetup() (map[string]any, error)
}

type Source struct {
	mux sync.Mutex
	id  uuid.UUID
	//address    net.Addr
	state     types.State
	sessionId uuid.UUID
	inputs    map[uuid.UUID]Input
	messages  chan Message
	inputCfgs map[uuid.UUID]inputConfig
}

type Message struct {
	Addr  net.Addr
	Event event.EventIface
}

func (s *Source) SetState(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return os.WriteFile("state.json", b, 0644)
}

func (s *Source) GetState(v any) error {
	b, err := os.ReadFile("state.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}

type inputConfig struct {
	cfg      map[string]any
	sinkCfgs []sinkConfig
}

type sinkConfig struct {
	Id      uuid.UUID
	Outputs []outputConfig
}

type outputConfig struct {
	Id     uuid.UUID
	Config map[string]any
	Leds   int
}

func New(id uuid.UUID) (*Source, error) {
	s := &Source{
		id: id,
		//address:   address,
		state:     types.StateIdle,
		inputs:    make(map[uuid.UUID]Input),
		messages:  make(chan Message),
		inputCfgs: map[uuid.UUID]inputConfig{},
	}

	return s, nil
}

func (s *Source) AddInput(i Input) {
	//fmt.Println("ADD INPUT CALLED", i)

	s.inputs[i.Id()] = i
	s.inputCfgs[i.Id()] = inputConfig{}

	go func() {
		// forward messages from input to the network
		for e := range i.Events() {
			var outputs []event.DataEventOutput
			for _, output := range e.Outputs {
				outputs = append(outputs, event.DataEventOutput{
					Id:  output.OutputId,
					Pix: output.Pix,
				})
			}

			s.messages <- Message{
				Addr: nil, // TODO: registry addr
				Event: event.DataEvent{
					Event:     event.Event{Type: event.Data},
					SessionId: s.sessionId,
					Outputs:   outputs,
				},
			}
		}
	}()
}

func (s *Source) RemoveInput(id uuid.UUID) {
	fmt.Println("RemoveInput CALLED", id)
	delete(s.inputs, id)
}

func (s *Source) Messages() <-chan Message {
	return s.messages
}

func (s *Source) ProcessEvent(addr net.Addr, e event.EventIface) {
	s.mux.Lock()
	defer s.mux.Unlock()

	switch e := e.(type) {
	case event.ConnectEvent:
		fmt.Printf("-> source %s: recv SetSourceActiveEvent\n", s.id)
		s.handleConnectedEvent(e)
	case event.SetSourceActiveEvent:
		fmt.Printf("-> source %s: recv SetSourceActiveEvent\n", s.id)
		s.handleSetActiveEvent(e)
	case event.SetSourceIdleEvent:
		fmt.Printf("-> source %s: recv SetSourceIdleEvent\n", s.id)
		s.handleSetIdleEvent(e)
	case event.ListCapabilitiesEvent:
		fmt.Printf("-> source %s: recv ListCapabilitiesEvent\n", s.id)
		s.handleListCapabilitiesEvent(addr, e)
	//case event.SetInputConfigEvent:
	//	fmt.Printf("-> source %s: recv SetInputConfigEvent\n", s.id)
	//	s.handleSetInputConfigEvent(e)
	case event.AssistedSetupEvent:
		fmt.Printf("-> source %s: recv AssistedSetupEvent\n", s.id)
		s.handleAssistedSetupEvent(addr, e)
	default:
		fmt.Println("unknrown event", e)
	}
}

func (s *Source) handleSetActiveEvent(e event.SetSourceActiveEvent) {
	if s.state == types.StateIdle {
		fmt.Println("Initializing session", e)

		s.state = types.StateActive
		s.sessionId = e.SessionId

		for _, input := range e.Inputs {
			cfg := s.inputCfgs[input.Id]
			cfg.sinkCfgs = lo.Map(input.Sinks, func(sink event.SetSourceActiveEventSink, index int) sinkConfig {
				return sinkConfig{
					Id: sink.Id,
					Outputs: lo.Map(sink.Outputs, func(output event.SetSourceActiveEventOutput, _ int) outputConfig {
						return outputConfig(output)
					}),
				}
			})

			// TODO: 1 config per output
			//err := s.inputs[input.Id].ApplyConfig(input.Config)
			//if err != nil {
			//	fmt.Println("error applying config", err)
			//	return
			//}
			//s.inputCfgs[input.OutputId].cfg = input.Config
			s.inputCfgs[input.Id] = cfg
		}

		var cfg types.InputConfig
		for inputId := range s.inputCfgs {
			// TODO: not the best solution to skip unrelated inputs
			if len(s.inputCfgs[inputId].sinkCfgs) == 0 {
				continue
			}

			for _, sinkCfg := range s.inputCfgs[inputId].sinkCfgs {
				for _, outputCfg := range sinkCfg.Outputs {
					cfg.Outputs = append(cfg.Outputs, types.OutputConfig{
						Id:     outputCfg.Id,
						SinkId: sinkCfg.Id,
						Config: outputCfg.Config,
						Leds:   outputCfg.Leds,
					})
				}
			}

			cfg.Framerate = 30

			_ = s.inputs[inputId].Start(cfg)
		}

		// TODO: validate there are no sourceIds present in the event but not present on the driver

		//for _, inputCfg := range e.Sources {
		//
		//	go func() {
		//		for {
		//			for _, sink := range s.inputCfgs[input.OutputId] {
		//				var outputs []event.DataEventOutput
		//				for _, output := range sink.Outputs {
		//					pix := make([]byte, output.Leds*4)
		//
		//					outputs = append(outputs, event.DataEventOutput{
		//						OutputId:  output.OutputId,
		//						Pix: pix,
		//					})
		//				}
		//
		//				s.messages <- event.DataEvent{
		//					Payload:     event.Payload{Type: event.Data, Addr: sink.OutputId},
		//					SessionId: s.sessionId,
		//					Outputs:   outputs,
		//				}
		//			}
		//
		//			time.Sleep(1 * time.Second)
		//		}
		//	}()
		//}

		//for _, input := range e.Sources {
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
		//	//		for ue := range input.Messages() {
		//	//			for _, sinkCfg := range sinkCfgs {
		//	//				e := regevent.SetLedsEvent{
		//	//					SessionId: uuid.Nil, // TODO
		//	//					OutputId:    sinkCfg.OutputId,
		//	//					// TODO: apply calibration to pix and pass clamped as byte array
		//	//					Pix: nil, // calibrate(ue.Pix)
		//	//				}
		//	//
		//	//				_ = ue
		//	//
		//	//				s.messages <- e
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

		for _, in := range s.inputs {
			err := in.Stop()
			if err != nil {
				fmt.Println("failed to stop source", err)
			}
		}
	}
}

func (s *Source) Id() uuid.UUID {
	return s.id
}

func (s *Source) handleListCapabilitiesEvent(addr net.Addr, e event.ListCapabilitiesEvent) {
	//fmt.Println("LISTING CAPABILITIES", fmt.Sprintf("%#v", s.inputs))
	s.messages <- Message{
		Addr: addr,
		Event: event.CapabilitiesEvent{
			Event: event.Event{Type: event.Capabilities},
			Id:    s.id,
			Inputs: lo.Map(lo.Values(s.inputs), func(input Input, _ int) event.CapabilitiesEventInput {
				return event.CapabilitiesEventInput{
					Id:           input.Id(),
					Type:         event.InputTypeDefault,
					ConfigSchema: input.Schema(),
				}
			}),
			Outputs: []event.CapabilitiesEventOutput{},
		},
	}
}

//func (s *Source) handleSetInputConfigEvent(e event.SetInputConfigEvent) {
//	input, ok := s.inputs[e.InputId]
//	if !ok {
//		fmt.Print("input not found")
//		return
//	}
//
//	err := input.ApplyConfig(e.Config)
//	if err != nil {
//		fmt.Println("failed to apply input config", err)
//		return
//	}
//}

func (s *Source) handleAssistedSetupEvent(addr net.Addr, e event.AssistedSetupEvent) {
	input, ok := s.inputs[e.InputId]
	if !ok {
		fmt.Print("input not found")
		return
	}

	cfg, err := input.AssistedSetup()
	if err != nil {
		fmt.Println("failed to get assisted setup", err)
		return
	}

	//inputCfg := s.inputCfgs[e.InputId]
	//inputCfg.cfg = cfg
	//s.inputCfgs[e.InputId] = inputCfg

	//err = s.inputs[e.InputId].ApplyConfig(cfg)
	//if err != nil {
	//	fmt.Println("failed to apply input config", err)
	//	return
	//}

	s.messages <- Message{
		Addr: addr,
		Event: event.AssistedSetupConfigEvent{
			Event:    event.Event{Type: event.Capabilities},
			SourceId: s.id,
			InputId:  e.InputId,
			Config:   cfg,
		},
	}
}

func (s *Source) handleConnectedEvent(e event.ConnectEvent) {

}
