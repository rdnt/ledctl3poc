package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"

	"ledctl3/event"
)

type Registry struct {
	id       uuid.UUID
	mux      sync.Mutex
	sources  map[uuid.UUID]Source
	sinks    map[uuid.UUID]Sink
	profiles map[uuid.UUID]Profile
	events   chan event.EventIface
}

type Profile struct {
	Id    uuid.UUID
	Name  string
	MapIO []map[uuid.UUID][]uuid.UUID
}

func New() *Registry {
	return &Registry{
		id:       uuid.New(),
		sources:  map[uuid.UUID]Source{},
		sinks:    map[uuid.UUID]Sink{},
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

type Source interface {
	Id() uuid.UUID
	Name() string
	String() string
	Inputs() map[uuid.UUID]Input
	Handle(event event.EventIface) error
	Events() <-chan event.EventIface
}

type InputState string

const (
	InputStateIdle   InputState = "idle"
	InputStateActive InputState = "active"
)

type OutputState string

const (
	OutputStateIdle   OutputState = "idle"
	OutputStateActive OutputState = "active"
)

type Input interface {
	Id() uuid.UUID
	Name() string
	State() InputState
	SessionId() uuid.UUID
}

type Calibration struct {
	R float64
	G float64
	B float64
	A float64
}

type Sink interface {
	Id() uuid.UUID
	Name() string
	Leds() int
	Calibration() map[int]Calibration
	Outputs() map[uuid.UUID]Output
	String() string
	Handle(event event.EventIface) error
	Events() <-chan event.EventIface
}

type Output interface {
	Id() uuid.UUID
	Name() string
	State() OutputState
	SessionId() uuid.UUID
	Leds() int
	Calibration() map[int]Calibration
}

var (
	ErrDeviceExists   = errors.New("device already exists")
	ErrDeviceNotFound = errors.New("device not found")
	ErrConfigNotFound = errors.New("config not found")
)

func (r *Registry) AddSource(src Source) error {
	_, ok := r.sources[src.Id()]
	if ok {
		return ErrDeviceExists
	}

	r.sources[src.Id()] = src

	return nil
}

func (r *Registry) AddSink(snk Sink) error {
	_, ok := r.sinks[snk.Id()]
	if ok {
		return ErrDeviceExists
	}

	r.sinks[snk.Id()] = snk

	return nil
}

func (r *Registry) Sources() map[uuid.UUID]Source {
	return r.sources
}

func (r *Registry) Sinks() map[uuid.UUID]Sink {
	return r.sinks
}

func (r *Registry) AddProfile(name string, mapIO []map[uuid.UUID][]uuid.UUID) Profile {
	cfg := Profile{
		Id:    uuid.New(),
		Name:  name,
		MapIO: mapIO,
	}

	r.profiles[cfg.Id] = cfg
	return cfg
}

func (r *Registry) SelectProfile(id uuid.UUID) error {
	cfg, ok := r.profiles[id]
	if !ok {
		return ErrConfigNotFound
	}

	sessId := uuid.New()

	var stopSessions []uuid.UUID
	enableSinkOutputs := map[uuid.UUID][]uuid.UUID{}

	fmt.Println("==== registry: configuring sink outputs")

	// enable sink outputs for this session
	for _, mapping := range cfg.MapIO {
		for _, outputIds := range mapping {
			for _, outputId := range outputIds {
				for _, snk := range r.sinks {
					output, ok := snk.Outputs()[outputId]
					if !ok {
						continue
					}

					stopSessions = append(stopSessions, output.SessionId())

					enableSinkOutputs[snk.Id()] = append(enableSinkOutputs[snk.Id()], outputId)
				}
			}
		}
	}

	for sinkId, outputIds := range enableSinkOutputs {
		err := r.sinks[sinkId].Handle(event.SetSinkActiveEvent{
			Event:     event.Event{Type: event.SetSinkActive, DevId: sinkId},
			SessionId: sessId,
			OutputIds: outputIds,
		})
		if err != nil {
			fmt.Println("error during send sink disabled outputs", err)
		}
	}

	//time.Sleep(1 * time.Second)
	fmt.Println("==== registry: disabling active source inputs")

	disableSourceInputs := map[uuid.UUID][]uuid.UUID{}

	// stop active inputs broadcasting to interrupted sessions
	for _, mapping := range cfg.MapIO {
		for inputId := range mapping {
			for _, src := range r.sources {
				input, ok := src.Inputs()[inputId]
				if !ok {
					continue
				}

				if input.State() == InputStateActive && slices.Contains(stopSessions, input.SessionId()) {
					// this input is broadcasting to a now-idle output. stop it
					disableSourceInputs[src.Id()] = append(disableSourceInputs[src.Id()], inputId)
				}

			}

		}
	}

	for srcId, inputIds := range disableSourceInputs {
		err := r.sources[srcId].Handle(event.SetSourceIdleEvent{
			Event: event.Event{Type: event.SetSourceIdle, DevId: srcId},
			//EventIface:    regevent.EventIface{EventDeviceId: sinkId},
			InputIds: inputIds,
		})
		if err != nil {
			fmt.Println("error during send source idle inputs", err)
		}
	}

	//time.Sleep(1 * time.Second)
	fmt.Println("==== registry: enabling inputs")

	//                     srcId  -->  inputIds  -->  sinkIds  -->  outputIds
	enableSourceIO := map[uuid.UUID]map[uuid.UUID]map[uuid.UUID][]uuid.UUID{}

	for _, mapping := range cfg.MapIO {
		for inputId, outputIds := range mapping {
			for _, src := range r.sources {
				_, ok := src.Inputs()[inputId]
				if !ok {
					continue
				}

				if _, ok := enableSourceIO[src.Id()]; !ok {
					enableSourceIO[src.Id()] = map[uuid.UUID]map[uuid.UUID][]uuid.UUID{}
				}

				if _, ok := enableSourceIO[src.Id()][inputId]; !ok {
					enableSourceIO[src.Id()][inputId] = map[uuid.UUID][]uuid.UUID{}
				}

				for _, outputId := range outputIds {

					for _, snk := range r.sinks {
						_, ok := snk.Outputs()[outputId]
						if !ok {
							continue
						}

						if _, ok := enableSourceIO[src.Id()][inputId][snk.Id()]; !ok {
							enableSourceIO[src.Id()][inputId][snk.Id()] = []uuid.UUID{}
						}

						enableSourceIO[src.Id()][inputId][snk.Id()] = append(enableSourceIO[src.Id()][inputId][snk.Id()], outputId)
					}
				}
			}
		}
	}

	//fmt.Println(lo.Values(lo.MapValues(r.sources, func(v Source, _ uuid.UUID) string {
	//	return fmt.Sprintf("\nsource %s (%s): %s", v.Id(), v.Name(), lo.Values(lo.MapValues(v.Inputs(), func(v Input, _ uuid.UUID) string {
	//		return fmt.Sprintf("\ninput %s (%s)", v.Id(), v.Name())
	//	})))
	//})))
	//
	//fmt.Println(lo.Values(lo.MapValues(r.sinks, func(v Sink, _ uuid.UUID) string {
	//	return fmt.Sprintf("\nsink %s (%s):\n%s", v.Id(), v.Name(), lo.Values(lo.MapValues(v.Outputs(), func(v Output, _ uuid.UUID) string {
	//		return fmt.Sprintf("\noutput %s (%s)", v.Id(), v.Name())
	//	})))
	//})))

	for srcId, inputSinkOutputs := range enableSourceIO {

		var inputs []event.SetSourceActiveEventInput

		for inputId, sinkOutputs := range inputSinkOutputs {

			var sinks []event.SetSourceActiveEventSink
			for sinkId, outputIds := range sinkOutputs {

				var outputs []event.SetSourceActiveEventOutput
				for _, outputId := range outputIds {
					output := r.sinks[sinkId].Outputs()[outputId]

					outputs = append(outputs, event.SetSourceActiveEventOutput{
						Id:   outputId,
						Leds: output.Leds(),
					})
				}

				sinks = append(sinks, event.SetSourceActiveEventSink{
					Id:      sinkId,
					Outputs: outputs,
				})
			}

			inputs = append(inputs, event.SetSourceActiveEventInput{
				Id:    inputId,
				Sinks: sinks,
			})
		}

		err := r.sources[srcId].Handle(event.SetSourceActiveEvent{
			Event:     event.Event{Type: event.SetSourceActive, DevId: srcId},
			SessionId: sessId,
			Inputs:    inputs,
		})
		if err != nil {
			fmt.Println("error during send source idle inputs", err)
		}
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
