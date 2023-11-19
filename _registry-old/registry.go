package _registry_old

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/samber/lo"

	"ledctl3/pkg/uuid"

	"ledctl3/_registry-old/types/sink"
	"ledctl3/_registry-old/types/source"
	"ledctl3/event"
)

type Store interface {
	Profiles() (map[uuid.UUID]Profile, error)
	SetProfiles(profiles map[uuid.UUID]Profile) error
	//Sources() (map[uuid.UUID]*source.Source, error)
	//SetSources(src map[uuid.UUID]*source.Source) error
}

type Registry struct {
	store    Store
	id       uuid.UUID
	mux      sync.Mutex
	sources  map[uuid.UUID]*source.Source
	sinks    map[uuid.UUID]*sink.Sink
	profiles map[uuid.UUID]Profile
	messages chan Message
}

type Message struct {
	Addr    net.Addr
	Payload event.EventIface
}

type Profile struct {
	Id      uuid.UUID
	Name    string
	Sources []ProfileSource
}

type ProfileSource struct {
	SourceId uuid.UUID
	Inputs   []ProfileInput
}

type ProfileInput struct {
	InputId uuid.UUID
	Sinks   []ProfileSink
}

type ProfileSink struct {
	SinkId  uuid.UUID
	Outputs []ProfileOutput
}

type ProfileOutput struct {
	OutputId      uuid.UUID
	InputConfigId uuid.UUID
}

func New(store Store) (*Registry, error) {
	r := &Registry{
		id:       uuid.New(),
		store:    store,
		sources:  map[uuid.UUID]*source.Source{},
		sinks:    map[uuid.UUID]*sink.Sink{},
		profiles: map[uuid.UUID]Profile{},
		messages: make(chan Message),
	}

	profs, err := store.Profiles()
	if err != nil {
		fmt.Println("get profiles: ", err)
		//return nil, fmt.Errorf("get profiles: %w", err)
	} else {
		r.profiles = profs
	}

	//sources, err := store.Sources()
	//if err != nil {
	//	fmt.Println("get sources: ", err)
	//	//return nil, fmt.Errorf("get sources: %w", err)
	//} else {
	//	r.sources = sources
	//}

	return r, nil
}

func (r *Registry) Id() uuid.UUID {
	return r.id
}

func (r *Registry) String() string {
	return fmt.Sprintf("registry{sources: %s, sinks: %s, profiles: %s}\n\n%#v", r.sources, r.sinks, r.profiles, r)
}

var (
	ErrDeviceExists   = errors.New("device already exists")
	ErrConfigNotFound = errors.New("config not found")
)

//func (r *Registry) RegisterDevice(id uuid.UUID, addr net.Addr) error {
//	err := r.AddSource(id, addr)
//	if err != nil {
//		return fmt.Errorf("add source: %w", err)
//	}
//
//	r.messages <- Message{
//		Addr: addr,
//		Payload: event.ListCapabilities{
//			Event: event.Event{Type: event.ListCapabilities},
//		},
//	}
//
//	return nil
//}

//func (r *Registry) AddSource(id uuid.UUID, addr net.Addr) error {
//	_, ok := r.sources[id]
//	if ok {
//		return ErrDeviceExists
//	}
//
//	src := source.New(id, addr)
//	r.sources[id] = src
//
//	//err := r.store.SetSources(r.sources)
//	//if err != nil {
//	//	fmt.Printf("error set sources: %s\n", err)
//	//	return err
//	//}
//
//	return nil
//}

//func (r *Registry) ConfigureInput(inputId uuid.UUID, cfg map[string]any) error {
//	for _, src := range r.sources {
//		for _, input := range src.Inputs {
//			if input.Id == inputId {
//				e := event.SetInputConfig{
//					Message:   event.Message{Type: event.SetInputConfig, Addr: src.Id},
//					Id: inputId,
//					Config:  cfg,
//				}
//
//				input, ok := r.sources[src.Id].Inputs[e.Id]
//				if !ok {
//					return fmt.Errorf("input %s not found", e.Id)
//				}
//
//				conf, err := input.AddConfig("", e.Config)
//				if err != nil {
//					return fmt.Errorf("apply input Config: %w", err)
//				}
//
//				err = input.ApplyConfig(conf.Id)
//				if err != nil {
//					return fmt.Errorf("apply input Config: %w", err)
//				}
//
//				r.messages <- e
//
//				return nil
//			}
//		}
//	}
//
//	return errors.New("input not found")
//}

func (r *Registry) AssistedSetup(inputId uuid.UUID) error {
	for _, src := range r.sources {
		for _, input := range src.Inputs {
			if input.Id == inputId {
				r.messages <- Message{
					Addr: src.Addr,
					Payload: event.AssistedSetup{
						Event:   event.Event{Type: event.AssistedSetup},
						InputId: inputId,
					},
				}

				return nil
			}
		}
	}

	return errors.New("input not found")
}

//func (r *Registry) AddSink(snk *sink.Sink) error {
//	_, ok := r.sinks[snk.Id]
//	if ok {
//		return ErrDeviceExists
//	}
//
//	r.sinks[snk.Id] = snk
//
//	return nil
//}

func (r *Registry) Sources() map[uuid.UUID]*source.Source {
	return r.sources
}

func (r *Registry) Sinks() map[uuid.UUID]*sink.Sink {
	return r.sinks
}

func (r *Registry) AddProfile(name string, sources []ProfileSource) (Profile, error) {
	prof := Profile{
		Id:      uuid.New(),
		Name:    name,
		Sources: sources,
	}

	r.profiles[prof.Id] = prof

	err := r.store.SetProfiles(r.profiles)
	if err != nil {
		return Profile{}, err
	}

	return prof, nil
}

func (r *Registry) configureOutputs(sessionId uuid.UUID, prof Profile) {
	fmt.Println("==== registry: configuring sink outputs")

	sinkEvents := map[uuid.UUID]event.SetSinkActive{}

	// enable sink outputs for this session
	for _, src := range prof.Sources {
		for _, in := range src.Inputs {
			for _, snk := range in.Sinks {
				evt, ok := sinkEvents[snk.SinkId]
				if !ok {
					evt = event.SetSinkActive{
						Event:     event.Event{Type: event.SetSinkActive},
						SessionId: sessionId,
						OutputIds: []uuid.UUID{},
					}
				}

				for _, out := range snk.Outputs {
					evt.OutputIds = append(sinkEvents[snk.SinkId].OutputIds, out.OutputId)
				}

				sinkEvents[snk.SinkId] = evt
			}
		}
	}

	for sinkId, e := range sinkEvents {
		snk := r.sinks[sinkId]
		r.messages <- Message{
			Addr:    snk.Addr,
			Payload: e,
		}
	}
}

func (r *Registry) disableActiveInputs(prof Profile) {
	fmt.Println("==== registry: disabling potentially active source inputs")

	sourceEvents := map[uuid.UUID]event.SetSourceIdle{}

	// stop active inputs broadcasting to interrupted sessions
	for _, src := range prof.Sources {
		evt, ok := sourceEvents[src.SourceId]
		if !ok {
			evt = event.SetSourceIdle{
				Event:  event.Event{Type: event.SetSinkActive},
				Inputs: []event.SetSourceIdleInput{},
			}
		}

		for _, in := range src.Inputs {
			outputIds := []uuid.UUID{}

			for _, snk := range in.Sinks {
				for _, out := range snk.Outputs {
					outputIds = append(outputIds, out.OutputId)
				}
			}

			evt.Inputs = append(evt.Inputs, event.SetSourceIdleInput{
				InputId:   in.InputId,
				OutputIds: outputIds,
			})
		}

		sourceEvents[src.SourceId] = evt
	}

	//for _, inputId := range e.InputIds {
	//	r.sources[srcId].Inputs[inputId].State = source.InputStateIdle
	//	r.sources[srcId].Inputs[inputId].SessionId = uuid.Nil
	//}

	for srcId, e := range sourceEvents {
		src := r.sources[srcId]
		r.messages <- Message{
			Addr:    src.Addr,
			Payload: e,
		}
	}
}

func (r *Registry) enableInputs(sessionId uuid.UUID, prof Profile) {
	fmt.Println("==== registry: enabling inputs")

	for _, src := range prof.Sources {
		//e := event.SetSourceActive{
		//	Message:     event.Message{Type: event.SetSourceActive, Addr: src.Id},
		//	SessionId: sessionId,
		//	Inputs: lo.Map(src.Inputs, func(in ProfileInput, _ int) event.SetSourceActiveInput {
		//		return event.SetSourceActiveInput{
		//			Id: in.Id,
		//			Sinks: lo.Map(in.Sinks, func(snk ProfileSink, _ int) event.SetSourceActiveSink {
		//				return event.SetSourceActiveSink{
		//					Id: snk.Id,
		//					Outputs: lo.Map(snk.Outputs, func(out ProfileOutput, _ int) event.SetSourceActiveOutput {
		//						return event.SetSourceActiveOutput{
		//							Id:     out.Id,
		//							Config: r.sources[src.Id].Inputs[in.Id].Configs[out.InputConfigId].Config,
		//							Leds:   r.sinks[snk.Id].Outputs[out.Id].Leds,
		//						}
		//					}),
		//				}
		//			}),
		//		}
		//	}),
		//}

		e := event.SetSourceActive{
			SessionId: sessionId,
			Inputs:    []event.SetSourceActiveInput{},
		}

		for _, in := range src.Inputs {
			inputCfg := event.SetSourceActiveInput{
				Id:    in.InputId,
				Sinks: []event.SetSourceActiveSink{},
			}

			for _, snk := range in.Sinks {
				sinkCfg := event.SetSourceActiveSink{
					Id:      snk.SinkId,
					Outputs: []event.SetSourceActiveOutput{},
				}

				for _, out := range snk.Outputs {
					sinkCfg.Outputs = append(sinkCfg.Outputs, event.SetSourceActiveOutput{
						Id:     out.OutputId,
						Config: r.sources[src.SourceId].Inputs[in.InputId].Configs[out.InputConfigId].Cfg,
						Leds:   r.sinks[snk.SinkId].Outputs[out.OutputId].Leds,
					})
				}

				inputCfg.Sinks = append(inputCfg.Sinks, sinkCfg)
			}

			e.Inputs = append(e.Inputs, inputCfg)
		}

		fmt.Println("@@@", fmt.Sprintf("%#v", e))
		src := r.sources[src.SourceId]
		r.messages <- Message{
			Addr:    src.Addr,
			Payload: e,
		}
	}
}

func (r *Registry) SelectProfile(id uuid.UUID, enable bool) error {
	prof, ok := r.profiles[id]
	if !ok {
		return ErrConfigNotFound
	}

	sessId := uuid.New()

	if enable {
		r.configureOutputs(sessId, prof)
	}

	time.Sleep(1 * time.Second)

	r.disableActiveInputs(prof)

	time.Sleep(1 * time.Second)

	if enable {
		r.enableInputs(sessId, prof)
	}

	return nil
}

func (r *Registry) Messages() <-chan Message {
	return r.messages
}

func (r *Registry) ProcessEvent(addr net.Addr, e event.EventIface) {
	r.mux.Lock()
	defer r.mux.Unlock()

	switch e := e.(type) {
	case event.Connect:
		fmt.Printf("%s -> registry: recv Connect\n", addr)
		r.handleConnectEvent(addr, e)
	//case event.SetSourceIdle:
	//	fmt.Printf("-> source %s: recv SetSourceIdle\n", s.id)
	//	s.handleSetIdleEvent(e)
	case event.Capabilities:
		fmt.Printf("%s -> registry: recv Capabilities\n", addr)
		r.handleCapabilitiesEvent(addr, e)
	case event.AssistedSetupConfig:
		fmt.Printf("%s -> registry: recv AssistedSetupConfig\n", addr)
		r.handleAssistedSetupConfigEvent(addr, e)

	default:
		fmt.Println("unknown event", e)
	}
}

func (r *Registry) handleConnectEvent(addr net.Addr, e event.Connect) {
	src, srcRegistered := r.sources[e.Id]
	snk, sinkRegistered := r.sinks[e.Id]

	if !srcRegistered || !sinkRegistered {
		fmt.Println("#### registry: unknown device", e.Id)

		r.messages <- Message{
			Addr: addr,
			Payload: event.ListCapabilities{
				Event: event.Event{Type: event.ListCapabilities},
			},
		}

		return
	}

	if srcRegistered {
		src.Addr = addr
		fmt.Println("Source address updated")
	}

	if sinkRegistered {
		snk.Addr = addr
		fmt.Println("Sink address updated")
	}
}

func (r *Registry) handleCapabilitiesEvent(addr net.Addr, e event.Capabilities) {
	if len(e.Inputs) > 0 {
		src, ok := r.sources[e.Id]
		if !ok {
			src = source.New(e.Id, addr)
			r.sources[e.Id] = src

			fmt.Println("Source added")
		}

		inputs := lo.Map(e.Inputs, func(input event.CapabilitiesInput, _ int) *source.Input {
			return source.NewInput(input.Id, "", input.Schema)
		})

		inputsMap := lo.SliceToMap(inputs, func(input *source.Input) (uuid.UUID, *source.Input) {
			return input.Id, input
		})

		src.Inputs = inputsMap
		fmt.Println("Source inputs updated")
	}

	if len(e.Outputs) > 0 {
		snk, ok := r.sinks[e.Id]
		if !ok {
			snk = sink.New(e.Id, addr)
			r.sinks[e.Id] = snk

			fmt.Println("Sink added")
		}

		outputs := lo.Map(e.Outputs, func(output event.CapabilitiesOutput, _ int) *sink.Output {
			return sink.NewOutput(output.Id, "", output.Leds)
		})

		outputsMap := lo.SliceToMap(outputs, func(output *sink.Output) (uuid.UUID, *sink.Output) {
			return output.Id, output
		})

		snk.Outputs = outputsMap
		fmt.Println("Sink outputs updated")
	}

	//err := r.store.SetSources(r.sources)
	//if err != nil {
	//	fmt.Printf("error set sources: %s\n", err)
	//}
}

func (r *Registry) handleAssistedSetupConfigEvent(addr net.Addr, e event.AssistedSetupConfig) {
	src, srcExists := r.sources[e.SourceId]
	_, snkExists := r.sinks[e.SourceId]

	if !srcExists && !snkExists {
		fmt.Println("registry: handleAssistedSetupConfigEvent: unknown device", e.SourceId)
		return
	}

	if srcExists {
		input, ok := r.sources[src.Id].Inputs[e.InputId]
		if !ok {
			fmt.Printf("input %s not found\n", e.InputId)
			return
		}

		_, err := input.AddConfig("", e.Config)
		if err != nil {
			fmt.Printf("apply input Config: %w\n", err)
			return
		}

	} else if snkExists {
		// TODO: will sinks need assisted setup?
	}
}

func (r *Registry) InputConfigs(inputId uuid.UUID) map[uuid.UUID]source.InputConfig {
	for _, src := range r.sources {
		if input, ok := src.Inputs[inputId]; ok {
			return input.Configs
		}

	}

	return nil
}

func (r *Registry) UpdateInputConfig(inputId, inputCfgId uuid.UUID, name string, cfg map[string]any) {
	for _, src := range r.sources {
		if input, ok := src.Inputs[inputId]; ok {
			input.Config = input.Configs[inputCfgId]

			input.Config.Name = name
			input.Config.Cfg = cfg

			input.Configs[inputCfgId] = input.Config
			break
		}
	}
}

//func (r *Registry) Connect(id uuid.UUID) {
//	//r.messages <- event.ListCapabilities{
//	//	Message: event.Message{Type: event.ListCapabilities, Addr: e.Id},
//	//}
//
//	//r.messages <- event.Connect{
//	//	Message: event.Message{Type: event.Connect, Addr: id},
//	//	Id:    r.id,
//	//}
//}
