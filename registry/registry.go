package registry

import (
	"encoding/gob"
	"errors"
	"fmt"
	"slices"
	"sync"

	"ledctl3/node/event"
	"ledctl3/pkg/uuid"
)

func init() {
	gob.Register(State{})
}

type StateHolder interface {
	SetState(state State) error
	GetState() (State, error)
}

type State struct {
	Nodes          map[uuid.UUID]*Node   `json:"nodes"`
	Profiles       map[uuid.UUID]Profile `json:"profiles"`
	ActiveProfiles []uuid.UUID           `json:"activeProfiles"`
}

type Registry struct {
	mux       sync.Mutex
	conns     map[string]uuid.UUID
	connsAddr map[uuid.UUID]string
	write     func(addr string, e event.Event) error
	request   func(addr string, e event.Event) error
	State     *State
	sh        StateHolder
	//handlers  map[uuid.UUID]map[uint64]func(string, event.Event) error
}

func New(sh StateHolder, write func(addr string, e event.Event) error, req func(addr string, e event.Event) error) *Registry {
	state, err := sh.GetState()
	if err == nil {
		//fmt.Println("Loaded State", State)
	} else {
		fmt.Println("error reading State", err)
		state = State{}
	}

	if state.Nodes == nil {
		state.Nodes = make(map[uuid.UUID]*Node)
	}

	if state.Profiles == nil {
		state.Profiles = make(map[uuid.UUID]Profile)
	}

	//fmt.Println("Starting with State", fmt.Sprintf("%#v", State))

	return &Registry{
		conns:     make(map[string]uuid.UUID),
		connsAddr: make(map[uuid.UUID]string),
		State:     &state,
		write:     write,
		request:   req,
		sh:        sh,
	}
}

type Profile struct {
	Id   uuid.UUID  `json:"id"`
	Name string     `json:"name"`
	IO   []IOConfig `json:"io"`
}

type IOConfig struct {
	InputId  uuid.UUID      `json:"input_id"`
	OutputId uuid.UUID      `json:"output_id"`
	Config   map[string]any `json:"config"`
}

//type Profile struct {
//	OutputId      uuid.UUID       `json:"id"`
//	Name    string          `json:"name"`
//	Sources []ProfileSource `json:"sources"`
//}
//
//type ProfileSource struct {
//	OutputId     uuid.UUID 	     `json:"id"`
//	Inputs []ProfileInput `json:"inputs"`
//}
//
//type ProfileInput struct {
//	OutputId    uuid.UUID     `json:"id"`
//	Sinks []ProfileSink `json:"sinks"`
//}
//
//type ProfileSink struct {
//	OutputId      uuid.UUID       `json:"id"`
//	Outputs []ProfileOutput `json:"outputs"`
//}
//
//type ProfileOutput struct {
//	OutputId            uuid.UUID `json:"id"`
//	InputConfigId uuid.UUID `json:"input_config_id"`
//}

var ErrEmptyIO = errors.New("empty io")

func (r *Registry) CreateProfile(name string, io []IOConfig) (Profile, error) {
	if len(io) == 0 {
		return Profile{}, ErrEmptyIO
	}

	prof := Profile{
		Id:   uuid.New(),
		Name: name,
		IO:   io,
	}

	r.State.Profiles[prof.Id] = prof

	err := r.sh.SetState(*r.State)
	if err != nil {
		return Profile{}, err
	}

	return prof, nil
}

func (r *Registry) SetSinkConfig(nodeId, sinkId uuid.UUID, cfg []byte) error {
	r.State.Nodes[nodeId].Sinks[sinkId].Config = cfg

	err := r.sh.SetState(*r.State)
	if err != nil {
		return err
	}

	err = r.send(r.connsAddr[nodeId], event.SetSinkConfig{
		SinkId: sinkId,
		Config: cfg,
	})
	if err != nil {
		fmt.Println("error sending event:", err)
		return err
	}

	return nil
}

func (r *Registry) EnableProfile(id uuid.UUID) error {
	prof, ok := r.State.Profiles[id]
	if !ok {
		return errors.New("profile not found")
	}

	if slices.Contains(r.State.ActiveProfiles, id) {
		return errors.New("profile already enabled")
	}

	activeOutputIds := r.activeOutputs()

	// cannot have multiple outputs active with different inputs.
	// maybe in the future, input data can be combined to the same
	// output with some transformer function, e.g. one input
	// modifying hue/sat and another modifying brightness.
	for _, io := range prof.IO {
		if slices.Contains(activeOutputIds, io.OutputId) {
			return errors.New("output already in use")
		}
	}

	r.State.ActiveProfiles = append(r.State.ActiveProfiles, id)

	err := r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	for _, io := range prof.IO {
		srcDev := r.State.Nodes[r.inputNodeId(io.InputId)]
		sinkDev := r.State.Nodes[r.outputNodeId(io.OutputId)]

		err = r.send(r.connsAddr[srcDev.Id], event.SetInputActive{
			Id: io.InputId,
			Outputs: []event.SetInputActiveOutput{
				{
					OutputId: io.OutputId,
					SinkId:   sinkDev.Id,
					Leds:     sinkDev.Outputs[io.OutputId].Leds,
					Config:   io.Config,
				},
			},
		})
		if err != nil {
			fmt.Println("error sending event:", err)
			continue
		}
	}

	fmt.Println("profile enabled:", id)
	return nil
}

func (r *Registry) activeOutputs() []uuid.UUID {
	var outputIds []uuid.UUID

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]
		for _, io := range prof.IO {
			outputIds = append(outputIds, io.OutputId)
		}
	}

	return outputIds
}

func (r *Registry) activeSourceOutputs(id uuid.UUID) []uuid.UUID {
	var outputIds []uuid.UUID

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]
		for _, io := range prof.IO {
			srcId := r.inputNodeId(io.InputId)

			if srcId != id {
				continue
			}

			outputIds = append(outputIds, io.OutputId)
		}
	}

	return outputIds
}

func (r *Registry) activeInputConfigs(id uuid.UUID) []IOConfig {
	var cfgs []IOConfig

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]

		for _, io := range prof.IO {
			if io.InputId != id {
				continue
			}

			cfgs = append(cfgs, io)
		}
	}

	return cfgs
}

func (r *Registry) outputNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, out := range dev.Outputs {
			if out.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Registry) inputNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, in := range dev.Inputs {
			if in.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}
