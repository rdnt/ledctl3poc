package registry

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/samber/lo"

	"ledctl3/pkg/uuid"

	"ledctl3/event"
	"ledctl3/registry/types/sink"
	"ledctl3/registry/types/source"
)

type Store interface {
	Profiles() (map[uuid.UUID]Profile, error)
	SetProfiles(profiles map[uuid.UUID]Profile) error
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

func (r *Registry) RegisterDevice(id uuid.UUID, addr net.Addr) error {
	err := r.AddSource(id, addr)
	if err != nil {
		return fmt.Errorf("add source: %w", err)
	}

	r.messages <- Message{
		Addr: addr,
		Payload: event.ListCapabilitiesEvent{
			Event: event.Event{Type: event.ListCapabilities},
		},
	}

	return nil
}

func (r *Registry) AddSource(id uuid.UUID, addr net.Addr) error {
	_, ok := r.sources[id]
	if ok {
		return ErrDeviceExists
	}

	src := source.New(id, addr)
	r.sources[id] = src

	return nil
}

//func (r *Registry) ConfigureInput(inputId uuid.UUID, cfg map[string]any) error {
//	for _, src := range r.sources {
//		for _, input := range src.Inputs {
//			if input.Id == inputId {
//				e := event.SetInputConfigEvent{
//					Message:   event.Message{Type: event.SetInputConfig, Addr: src.Id},
//					InputId: inputId,
//					Config:  cfg,
//				}
//
//				input, ok := r.sources[src.Id].Inputs[e.InputId]
//				if !ok {
//					return fmt.Errorf("input %s not found", e.InputId)
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
					Payload: event.AssistedSetupEvent{
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

func (r *Registry) AddSink(snk *sink.Sink) error {
	_, ok := r.sinks[snk.Id]
	if ok {
		return ErrDeviceExists
	}

	r.sinks[snk.Id] = snk

	return nil
}

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

	sinkEvents := map[uuid.UUID]event.SetSinkActiveEvent{}

	// enable sink outputs for this session
	for _, src := range prof.Sources {
		for _, in := range src.Inputs {
			for _, snk := range in.Sinks {
				evt, ok := sinkEvents[snk.SinkId]
				if !ok {
					evt = event.SetSinkActiveEvent{
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

	sourceEvents := map[uuid.UUID]event.SetSourceIdleEvent{}

	// stop active inputs broadcasting to interrupted sessions
	for _, src := range prof.Sources {
		evt, ok := sourceEvents[src.SourceId]
		if !ok {
			evt = event.SetSourceIdleEvent{
				Event:  event.Event{Type: event.SetSinkActive},
				Inputs: []event.SetSourceIdleEventInput{},
			}
		}

		for _, in := range src.Inputs {
			outputIds := []uuid.UUID{}

			for _, snk := range in.Sinks {
				for _, out := range snk.Outputs {
					outputIds = append(outputIds, out.OutputId)
				}
			}

			evt.Inputs = append(evt.Inputs, event.SetSourceIdleEventInput{
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
		src := r.sinks[srcId]
		r.messages <- Message{
			Addr:    src.Addr,
			Payload: e,
		}
	}
}

func (r *Registry) enableInputs(sessionId uuid.UUID, prof Profile) {
	fmt.Println("==== registry: enabling inputs")

	for _, src := range prof.Sources {
		//e := event.SetSourceActiveEvent{
		//	Message:     event.Message{Type: event.SetSourceActive, Addr: src.SourceId},
		//	SessionId: sessionId,
		//	Inputs: lo.Map(src.Inputs, func(in ProfileInput, _ int) event.SetSourceActiveEventInput {
		//		return event.SetSourceActiveEventInput{
		//			Id: in.InputId,
		//			Sinks: lo.Map(in.Sinks, func(snk ProfileSink, _ int) event.SetSourceActiveEventSink {
		//				return event.SetSourceActiveEventSink{
		//					Id: snk.SinkId,
		//					Outputs: lo.Map(snk.Outputs, func(out ProfileOutput, _ int) event.SetSourceActiveEventOutput {
		//						return event.SetSourceActiveEventOutput{
		//							Id:     out.OutputId,
		//							Config: r.sources[src.SourceId].Inputs[in.InputId].Configs[out.InputConfigId].Config,
		//							Leds:   r.sinks[snk.SinkId].Outputs[out.OutputId].Leds,
		//						}
		//					}),
		//				}
		//			}),
		//		}
		//	}),
		//}

		e := event.SetSourceActiveEvent{
			Event:     event.Event{Type: event.SetSourceActive},
			SessionId: sessionId,
			Inputs:    []event.SetSourceActiveEventInput{},
		}

		for _, in := range src.Inputs {
			inputCfg := event.SetSourceActiveEventInput{
				Id:    in.InputId,
				Sinks: []event.SetSourceActiveEventSink{},
			}

			for _, snk := range in.Sinks {
				sinkCfg := event.SetSourceActiveEventSink{
					Id:      snk.SinkId,
					Outputs: []event.SetSourceActiveEventOutput{},
				}

				for _, out := range snk.Outputs {
					sinkCfg.Outputs = append(sinkCfg.Outputs, event.SetSourceActiveEventOutput{
						Id:     out.OutputId,
						Config: r.sources[src.SourceId].Inputs[in.InputId].Configs[out.InputConfigId].Cfg,
						Leds:   r.sinks[snk.SinkId].Outputs[out.OutputId].Leds,
					})
				}

				inputCfg.Sinks = append(inputCfg.Sinks, sinkCfg)
			}

			e.Inputs = append(e.Inputs, inputCfg)
		}

		src := r.sinks[src.SourceId]
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
	case event.ConnectEvent:
		fmt.Printf("%s -> registry: recv ConnectEvent\n", addr)
		r.handleConnectEvent(addr, e)
	//case event.SetSourceIdleEvent:
	//	fmt.Printf("-> source %s: recv SetSourceIdleEvent\n", s.id)
	//	s.handleSetIdleEvent(e)
	case event.CapabilitiesEvent:
		fmt.Printf("%s -> registry: recv CapabilitiesEvent\n", addr)
		r.handleCapabilitiesEvent(e)
	case event.AssistedSetupConfigEvent:
		fmt.Printf("%s -> registry %s: recv AssistedSetupConfigEvent\n", addr)
		r.handleAssistedSetupConfigEvent(e)

	default:
		fmt.Println("unknown event", e)
	}
}

func (r *Registry) handleConnectEvent(addr net.Addr, e event.ConnectEvent) {
	_, srcRegistered := r.sources[e.Id]
	_, sinkRegistered := r.sinks[e.Id]

	if !srcRegistered || !sinkRegistered {
		fmt.Println("#### registry: unknown device", e.Id)

		r.messages <- Message{
			Addr: addr,
			Payload: event.ListCapabilitiesEvent{
				Event: event.Event{Type: event.ListCapabilities},
			},
		}

		return
	}
}

func (r *Registry) handleCapabilitiesEvent(e event.CapabilitiesEvent) {
	if len(e.Inputs) > 0 {
		src, ok := r.sources[e.Id]
		if !ok {
			src = source.New(e.Id, nil)
			r.sources[e.Id] = src
		}

		inputs := lo.Map(e.Inputs, func(input event.CapabilitiesEventInput, _ int) *source.Input {
			return source.NewInput(input.Id, "", input.ConfigSchema)
		})

		inputsMap := lo.SliceToMap(inputs, func(input *source.Input) (uuid.UUID, *source.Input) {
			return input.Id, input
		})

		src.SetInputs(inputsMap)
	}

	if len(e.Outputs) > 0 {
		snk, ok := r.sinks[e.Id]
		if !ok {
			snk = sink.New(e.Id)
			r.sinks[e.Id] = snk
		}

		outputs := lo.Map(e.Outputs, func(output event.CapabilitiesEventOutput, _ int) *sink.Output {
			return sink.NewOutput(output.Id, "", output.Leds)
		})

		outputsMap := lo.SliceToMap(outputs, func(output *sink.Output) (uuid.UUID, *sink.Output) {
			return output.Id, output
		})

		snk.SetOutputs(outputsMap)
	}
}

func (r *Registry) handleAssistedSetupConfigEvent(e event.AssistedSetupConfigEvent) {
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
//	//r.messages <- event.ListCapabilitiesEvent{
//	//	Message: event.Message{Type: event.ListCapabilities, Addr: e.Id},
//	//}
//
//	//r.messages <- event.ConnectEvent{
//	//	Message: event.Message{Type: event.Connect, Addr: id},
//	//	Id:    r.id,
//	//}
//}
