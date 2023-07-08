package registry

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"

	"ledctl3/pkg/event"
	regevent "ledctl3/registry/types/event"
)

type Registry struct {
	sources  map[uuid.UUID]Source
	sinks    map[uuid.UUID]Sink
	profiles map[uuid.UUID]Profile
}

type Profile struct {
	Id    uuid.UUID
	Name  string
	MapIO []map[uuid.UUID][]uuid.UUID
}

func New() *Registry {
	return &Registry{
		sources:  map[uuid.UUID]Source{},
		sinks:    map[uuid.UUID]Sink{},
		profiles: map[uuid.UUID]Profile{},
	}
}

func (r *Registry) String() string {
	return fmt.Sprintf("registry{sources: %s, sinks: %s, profiles: %s}\n\n%#v", r.sources, r.sinks, r.profiles, r)
}

type Source interface {
	Id() uuid.UUID
	Name() string
	String() string
	Inputs() map[uuid.UUID]Input
	Handle(event event.Event) error
	Events() <-chan event.Event
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
	Handle(event event.Event) error
	Events() <-chan event.Event
}

type Output interface {
	Id() uuid.UUID
	Name() string
	State() OutputState
	SessionId() uuid.UUID
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

	fmt.Println("--------------- START ---------------")
	fmt.Println("=== registry: configure sink outputs")

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
		err := r.sinks[sinkId].Handle(regevent.SetSinkActiveEvent{
			Event:     regevent.Event{Type: regevent.SetSinkActive, DevId: sinkId},
			SessionId: sessId,
			OutputIds: outputIds,
		})
		if err != nil {
			fmt.Println("error during send sink disabled outputs", err)
		}
	}

	fmt.Println("----------------------------------------")
	time.Sleep(1 * time.Second)
	fmt.Println("----------------------------------------")
	fmt.Println("=== registry: disable active source inputs")

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
		err := r.sources[srcId].Handle(regevent.SetSourceIdleEvent{
			Event: regevent.Event{Type: regevent.SetSourceIdle, DevId: srcId},
			//Event:    regevent.Event{EventDeviceId: sinkId},
			InputIds: inputIds,
		})
		if err != nil {
			fmt.Println("error during send source idle inputs", err)
		}
	}

	fmt.Println("----------------------------------------")
	time.Sleep(1 * time.Second)
	fmt.Println("----------------------------------------")
	fmt.Println("=== registry: enable inputs")

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

						enableSourceIO[src.Id()][inputId][snk.Id()] = outputIds
					}
				}
			}
		}
	}

	for srcId, inputSinkOutputs := range enableSourceIO {

		sinkCfgs := map[uuid.UUID][]regevent.SetSourceActiveEventSink{}

		for inputId, sinkOutputs := range inputSinkOutputs {
			for sinkId, outputIds := range sinkOutputs {
				sinkCfgs[inputId] = append(sinkCfgs[inputId], regevent.SetSourceActiveEventSink{
					Id:        sinkId,
					Address:   "", // TODO ?
					OutputIds: outputIds,
				})
			}
		}

		err := r.sources[srcId].Handle(regevent.SetSourceActiveEvent{
			Event:     regevent.Event{Type: regevent.SetSourceActive, DevId: srcId},
			SessionId: sessId,
			Sinks:     sinkCfgs,
		})
		if err != nil {
			fmt.Println("error during send source idle inputs", err)
		}
	}

	return nil
}

//func (r *Registry) setState(srcId uuid.UUID, devId uuid.UUID, state types.State) error {
//	_, ok := r.sources[srcId]
//	if !ok {
//		return ErrDeviceNotFound
//	}
//
//	_, ok = r.sinks[devId]
//	if !ok {
//		return ErrDeviceNotFound
//	}
//
//	if state == types.StateActive {
//		sessId := uuid.New()
//
//		// set sinks to active for new session
//		for _, s := range r.sinks {
//			err := s.Handle(regevent.SetSinkActiveEvent{
//				SessionId: sessId,
//				Leds:      0,
//			})
//			if err != nil {
//				fmt.Println("error during send sink active", err)
//			}
//		}
//
//		// go -> stop active sources TODO: relevant to this sink (dont stop all...)
//		for _, s := range r.sources {
//			if s.State() == source.StateActive {
//				err := s.Handle(regevent.SetSourceIdleEvent{})
//				if err != nil {
//					fmt.Println("error during send source idle", err)
//				}
//			}
//		}
//
//		// set sources to active for new session
//		for _, s := range r.sources {
//			if s.State() == source.StateActive {
//				err := s.Handle(regevent.SetSourceIdleEvent{})
//				if err != nil {
//					fmt.Println("error during send source idle", err)
//				}
//			}
//		}
//	}
//
//	// TODO: somehow we should notify sinks to start handling events for sessionId
//	//go func() {
//	//	for e := range r.sources[srcId].Events() {
//	//		r.sinks[devId].Handle(e)
//	//	}
//	//}()
//
//	// TODO fix
//	//// prepare the server to start receiving the events
//	//r.sinks[devId].SetState(state)
//	//
//	//// switch state on the source to start event transmission
//	//r.sources[srcId].SetState(state)
//
//	return nil
//}

//func (r *Registry) Events() <-chan event.Event {
//	return r.events
//}

func (r *Registry) ProcessEvent(e event.Event) {
	fmt.Println("UNHANDLED PROCESS REGISTRY", e)
}
