package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"

	"ledctl3/event"
	"ledctl3/registry/types/sink"
	"ledctl3/registry/types/source"
)

type Registry struct {
	id       uuid.UUID
	mux      sync.Mutex
	sources  map[uuid.UUID]*source.Source
	sinks    map[uuid.UUID]*sink.Sink
	profiles map[uuid.UUID]Profile
	events   chan event.EventIface
}

type Profile struct {
	Id     uuid.UUID
	Name   string
	Inputs []ProfileInput
}

type ProfileInput struct {
	InputId   uuid.UUID
	CfgId     uuid.UUID
	OutputIds []uuid.UUID
}

func New() *Registry {
	return &Registry{
		id:       uuid.New(),
		sources:  map[uuid.UUID]*source.Source{},
		sinks:    map[uuid.UUID]*sink.Sink{},
		profiles: map[uuid.UUID]Profile{},
		events:   make(chan event.EventIface),
	}
}

func (r *Registry) Id() uuid.UUID {
	return r.id
}

func (r *Registry) String() string {
	return fmt.Sprintf("registry{sources: %s, sinks: %s, profiles: %s}\n\n%#v", r.sources, r.sinks, r.profiles, r)
}

var (
	ErrDeviceExists   = errors.New("device already exists")
	ErrDeviceNotFound = errors.New("device not found")
	ErrConfigNotFound = errors.New("config not found")
)

func (r *Registry) AddSource(src *source.Source) error {
	_, ok := r.sources[src.Id]
	if ok {
		return ErrDeviceExists
	}

	r.sources[src.Id] = src

	return nil
}

func (r *Registry) ConfigureInput(inputId uuid.UUID, cfg map[string]any) error {
	for _, src := range r.sources {
		for _, input := range src.Inputs {
			if input.Id == inputId {
				e := event.SetInputConfigEvent{
					Event:   event.Event{Type: event.SetInputConfig, DevId: src.Id},
					InputId: inputId,
					Config:  cfg,
				}

				input, ok := r.sources[src.Id].Inputs[e.InputId]
				if !ok {
					return fmt.Errorf("input %s not found", e.InputId)
				}

				conf, err := input.AddConfig("", e.Config)
				if err != nil {
					return fmt.Errorf("apply input Config: %w", err)
				}

				err = input.ApplyConfig(conf.Id)
				if err != nil {
					return fmt.Errorf("apply input Config: %w", err)
				}

				r.events <- e

				return nil
			}
		}
	}

	return errors.New("input not found")
}

func (r *Registry) AssistedSetup(inputId uuid.UUID) error {
	for _, src := range r.sources {
		for _, input := range src.Inputs {
			if input.Id == inputId {
				e := event.AssistedSetupEvent{
					Event:   event.Event{Type: event.AssistedSetup, DevId: src.Id},
					InputId: inputId,
				}

				r.events <- e

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

func (r *Registry) AddProfile(name string, inputs []ProfileInput) Profile {
	prof := Profile{
		Id:     uuid.New(),
		Name:   name,
		Inputs: inputs,
	}

	r.profiles[prof.Id] = prof
	return prof
}

// TODO: cleanup this mess C:
func (r *Registry) SelectProfile(id uuid.UUID) error {
	prof, ok := r.profiles[id]
	if !ok {
		return ErrConfigNotFound
	}

	sessId := uuid.New()

	var stopSessions []uuid.UUID
	enableSinkOutputs := map[uuid.UUID][]uuid.UUID{}

	fmt.Println("==== registry: configuring sink outputs")

	// enable sink outputs for this session
	for _, inputProfile := range prof.Inputs {
		for _, outputId := range inputProfile.OutputIds {
			for _, snk := range r.sinks {
				output, ok := snk.Outputs[outputId]
				if !ok {
					continue
				}

				stopSessions = append(stopSessions, output.SessionId)

				enableSinkOutputs[snk.Id] = append(enableSinkOutputs[snk.Id], outputId)
			}
		}
	}

	for sinkId, outputIds := range enableSinkOutputs {
		e := event.SetSinkActiveEvent{
			Event:     event.Event{Type: event.SetSinkActive, DevId: sinkId},
			SessionId: sessId,
			OutputIds: outputIds,
		}

		r.sinks[sinkId].Process(e)
		r.events <- e
	}

	//time.Sleep(1 * time.Second)
	fmt.Println("==== registry: disabling active source inputs")

	disableSourceInputs := map[uuid.UUID][]uuid.UUID{}

	// stop active inputs broadcasting to interrupted sessions
	for _, inputProfile := range prof.Inputs {
		for _, src := range r.sources {
			input, ok := src.Inputs[inputProfile.InputId]
			if !ok {
				continue
			}

			if input.State == source.InputStateActive && slices.Contains(stopSessions, input.SessionId) {
				// this input is broadcasting to a now-idle output. stop it
				disableSourceInputs[src.Id] = append(disableSourceInputs[src.Id], inputProfile.InputId)
			}
		}
	}

	for srcId, inputIds := range disableSourceInputs {
		e := event.SetSourceIdleEvent{
			Event:    event.Event{Type: event.SetSourceIdle, DevId: srcId},
			InputIds: inputIds,
		}

		for _, inputId := range e.InputIds {
			r.sources[srcId].Inputs[inputId].State = source.InputStateIdle
			r.sources[srcId].Inputs[inputId].SessionId = uuid.Nil
		}

		r.events <- e
	}

	//time.Sleep(1 * time.Second)
	fmt.Println("==== registry: enabling inputs")

	//                     srcId  -->  inputIds  -->  sinkIds  -->  outputIds
	enableSourceIO := map[uuid.UUID]map[uuid.UUID]map[uuid.UUID][]uuid.UUID{}

	for _, inputProfile := range prof.Inputs {
		for _, outputId := range inputProfile.OutputIds {
			for _, src := range r.sources {
				_, ok := src.Inputs[inputProfile.InputId]
				if !ok {
					continue
				}

				if _, ok := enableSourceIO[src.Id]; !ok {
					enableSourceIO[src.Id] = map[uuid.UUID]map[uuid.UUID][]uuid.UUID{}
				}

				if _, ok := enableSourceIO[src.Id][inputProfile.InputId]; !ok {
					enableSourceIO[src.Id][inputProfile.InputId] = map[uuid.UUID][]uuid.UUID{}
				}

				for _, snk := range r.sinks {
					_, ok := snk.Outputs[outputId]
					if !ok {
						continue
					}

					if _, ok := enableSourceIO[src.Id][inputProfile.InputId][snk.Id]; !ok {
						enableSourceIO[src.Id][inputProfile.InputId][snk.Id] = []uuid.UUID{}
					}

					enableSourceIO[src.Id][inputProfile.InputId][snk.Id] = append(enableSourceIO[src.Id][inputProfile.InputId][snk.Id], outputId)
				}
			}
		}
	}

	//fmt.Println(lo.Values(lo.MapValues(r.sources, func(v Source, _ uuid.UUID) string {
	//	return fmt.Sprintf("\nsource %s (%s): %s", v.OutputId(), v.Name(), lo.Values(lo.MapValues(v.Inputs(), func(v Input, _ uuid.UUID) string {
	//		return fmt.Sprintf("\ninput %s (%s)", v.OutputId(), v.Name())
	//	})))
	//})))
	//
	//fmt.Println(lo.Values(lo.MapValues(r.sinks, func(v Sink, _ uuid.UUID) string {
	//	return fmt.Sprintf("\nsink %s (%s):\n%s", v.OutputId(), v.Name(), lo.Values(lo.MapValues(v.Outputs(), func(v Output, _ uuid.UUID) string {
	//		return fmt.Sprintf("\noutput %s (%s)", v.OutputId(), v.Name())
	//	})))
	//})))

	for srcId, inputSinkOutputs := range enableSourceIO {

		var inputs []event.SetSourceActiveEventInput

		for inputId, sinkOutputs := range inputSinkOutputs {

			var sinks []event.SetSourceActiveEventSink
			for sinkId, outputIds := range sinkOutputs {

				var outputs []event.SetSourceActiveEventOutput
				for _, outputId := range outputIds {
					output := r.sinks[sinkId].Outputs[outputId]

					outputs = append(outputs, event.SetSourceActiveEventOutput{
						Id:   outputId,
						Leds: output.Leds,
					})
				}

				sinks = append(sinks, event.SetSourceActiveEventSink{
					Id:      sinkId,
					Outputs: outputs,
				})
			}

			_, ok := r.sources[srcId].Inputs[inputId]
			if !ok {
				continue
			}

			var cfg map[string]any
			for _, inputProfile := range prof.Inputs {
				if inputProfile.InputId == inputId {
					cfg = r.sources[srcId].Inputs[inputId].Configs[inputProfile.CfgId].Cfg
				}
			}

			inputs = append(inputs, event.SetSourceActiveEventInput{
				Id:     inputId,
				Sinks:  sinks,
				Config: cfg,
			})
		}

		e := event.SetSourceActiveEvent{
			Event:     event.Event{Type: event.SetSourceActive, DevId: srcId},
			SessionId: sessId,
			Inputs:    inputs,
		}

		//for _, input := range e.Inputs {

		for _, input := range e.Inputs {
			in := r.sources[srcId].Inputs[input.Id]

			in.State = source.InputStateActive
			in.SessionId = e.SessionId
			in.Sinks = lo.Map(input.Sinks, func(sink event.SetSourceActiveEventSink, _ int) source.SinkConfig {
				return source.SinkConfig{
					Id: sink.Id,
					Outputs: lo.Map(sink.Outputs, func(output event.SetSourceActiveEventOutput, _ int) source.OutputConfig {
						return source.OutputConfig(output)
					}),
				}
			})

			//
			//var sinks []sink
			//for _, Config := range input.Sinks {
			//	s := sink{
			//		OutputId: Config.OutputId,
			//	}
			//
			//	for _, out := range Config.Outputs {
			//		s.outputs = append(s.outputs, output{
			//			OutputId:   out.OutputId,
			//			leds: out.Leds,
			//		})
			//	}
			//
			//	sinks = append(sinks, s)
			//}
		}

		r.events <- e
	}

	return nil
}

func (r *Registry) Events() <-chan event.EventIface {
	return r.events
}

func (r *Registry) ProcessEvent(e event.EventIface) {
	r.mux.Lock()
	defer r.mux.Unlock()

	switch e := e.(type) {
	case event.ConnectEvent:
		fmt.Printf("-> registry %s: recv ConnectEvent\n", r.id)
		r.handleConnectEvent(e)
	//case event.SetSourceIdleEvent:
	//	fmt.Printf("-> source %s: recv SetSourceIdleEvent\n", s.id)
	//	s.handleSetIdleEvent(e)
	case event.CapabilitiesEvent:
		fmt.Printf("-> registry %s: recv CapabilitiesEvent\n", r.id)
		r.handleCapabilitiesEvent(e)
	case event.AssistedSetupConfigEvent:
		fmt.Printf("-> registry %s: recv AssistedSetupConfigEvent\n", r.id)
		r.handleAssistedSetupConfigEvent(e)

	default:
		fmt.Println("unknown event", e)
	}
}

func (r *Registry) handleConnectEvent(e event.ConnectEvent) {
	_, srcRegistered := r.sources[e.Id]
	_, sinkRegistered := r.sinks[e.Id]

	if !srcRegistered || !sinkRegistered {
		fmt.Println("#### registry: unknown device", e.Id)

		r.events <- event.ListCapabilitiesEvent{
			Event: event.Event{Type: event.ListCapabilities, DevId: e.Id},
		}

		return
	}
}

func (r *Registry) handleCapabilitiesEvent(e event.CapabilitiesEvent) {
	if len(e.Inputs) > 0 {
		_, srcRegistered := r.sources[e.Id]

		if !srcRegistered {
			inputs := lo.Map(e.Inputs, func(input event.CapabilitiesEventInput, _ int) *source.Input {
				return source.NewInput(input.Id, "", input.ConfigSchema)
			})

			inputsMap := lo.SliceToMap(inputs, func(input *source.Input) (uuid.UUID, *source.Input) {
				return input.Id, input
			})

			src := source.NewSource(e.Id, "", inputsMap)

			err := r.AddSource(src)
			if err != nil {
				fmt.Println("error during add source", err)
				return
			}
		}
	}

	if len(e.Outputs) > 0 {
		_, sinkRegistered := r.sinks[e.Id]

		if !sinkRegistered {
			outputs := lo.Map(e.Outputs, func(output event.CapabilitiesEventOutput, _ int) *sink.Output {
				return sink.NewOutput(output.Id, "", output.Leds)
			})

			outputsMap := lo.SliceToMap(outputs, func(output *sink.Output) (uuid.UUID, *sink.Output) {
				return output.Id, output
			})

			snk := sink.NewSink(e.Id, "", outputsMap)

			err := r.AddSink(snk)
			if err != nil {
				fmt.Println("error during add sink", err)
				return
			}
		}
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
