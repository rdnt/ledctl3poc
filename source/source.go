package source

import (
	"fmt"
	"sync"

	"github.com/samber/lo"

	"ledctl3/pkg/uuid"

	"ledctl3/event"
	"ledctl3/source/types"
)

type Input interface {
	Id() uuid.UUID
	Start(cfg types.SinkConfig) error
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
	state      types.State
	sessionId  uuid.UUID
	inputs     map[uuid.UUID]Input
	events     chan event.EventIface
	inputCfgs  map[uuid.UUID]inputConfig
	registryId uuid.UUID
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

func New(registryId uuid.UUID) (*Source, error) {
	s := &Source{
		id: uuid.New(),
		//address:   address,
		state:      types.StateIdle,
		inputs:     make(map[uuid.UUID]Input),
		events:     make(chan event.EventIface),
		inputCfgs:  map[uuid.UUID]inputConfig{},
		registryId: registryId,
	}

	return s, nil
}

func (s *Source) AddInput(i Input) {
	s.inputs[i.Id()] = i
	s.inputCfgs[i.Id()] = inputConfig{}

	go func() {
		// forward events from input to the network
		for e := range i.Events() {
			var outputs []event.DataEventOutput
			for _, output := range e.Outputs {
				outputs = append(outputs, event.DataEventOutput{
					Id:  output.OutputId,
					Pix: output.Pix,
				})
			}

			s.events <- event.DataEvent{
				Event:     event.Event{Type: event.Data, DevId: e.SinkId},
				SessionId: s.sessionId,
				Outputs:   outputs,
			}
		}
	}()
}

func (s *Source) RemoveInput(id uuid.UUID) {
	delete(s.inputs, id)
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
	case event.ListCapabilitiesEvent:
		fmt.Printf("-> source %s: recv ListCapabilitiesEvent\n", s.id)
		s.handleListCapabilitiesEvent(e)
	//case event.SetInputConfigEvent:
	//	fmt.Printf("-> source %s: recv SetInputConfigEvent\n", s.id)
	//	s.handleSetInputConfigEvent(e)
	case event.AssistedSetupEvent:
		fmt.Printf("-> source %s: recv AssistedSetupEvent\n", s.id)
		s.handleAssistedSetupEvent(e)
	default:
		fmt.Println("unknrown event", e)
	}
}

func (s *Source) handleSetActiveEvent(e event.SetSourceActiveEvent) {
	if s.state == types.StateIdle {
		//fmt.Println("Initializing session", e.SessionId, e.Sinks)

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

		var cfg types.SinkConfig
		for inputId := range s.inputCfgs {
			// TODO: not the best solution to skip unrelated inputs
			if len(s.inputCfgs[inputId].sinkCfgs) == 0 {
				continue
			}

			for _, sinkCfg := range s.inputCfgs[inputId].sinkCfgs {

				var outputs []types.SinkConfigSinkOutput
				for _, outputCfg := range sinkCfg.Outputs {
					outputs = append(outputs, types.SinkConfigSinkOutput{
						Id:     outputCfg.Id,
						Config: outputCfg.Config,
						Leds:   outputCfg.Leds,
					})
				}

				cfg.Sinks = append(cfg.Sinks, types.SinkConfigSink{
					Id:      sinkCfg.Id,
					Outputs: outputs,
				})
			}

			cfg.Framerate = 60

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
		//				s.events <- event.DataEvent{
		//					Event:     event.Event{Type: event.Data, DevId: sink.OutputId},
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
		//	//		for ue := range input.Events() {
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

func (s *Source) handleListCapabilitiesEvent(_ event.ListCapabilitiesEvent) {
	s.events <- event.CapabilitiesEvent{
		Event: event.Event{Type: event.Capabilities, DevId: s.registryId},
		Id:    s.id,
		Inputs: lo.Map(lo.Values(s.inputs), func(input Input, _ int) event.CapabilitiesEventInput {
			return event.CapabilitiesEventInput{
				Id:           input.Id(),
				Type:         event.InputTypeDefault,
				ConfigSchema: input.Schema(),
			}
		}),
		Outputs: []event.CapabilitiesEventOutput{},
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

func (s *Source) handleAssistedSetupEvent(e event.AssistedSetupEvent) {
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

	s.events <- event.AssistedSetupConfigEvent{
		Event:    event.Event{Type: event.Capabilities, DevId: s.registryId},
		SourceId: s.id,
		InputId:  e.InputId,
		Config:   cfg,
	}
}
