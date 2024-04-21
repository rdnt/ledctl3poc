package orchestrator

import (
	"encoding/gob"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/google/uuid"

	"ledctl3/node/event"
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

type Orchestrator struct {
	mux sync.Mutex

	Conns map[string]Conn

	AddrNode map[string]uuid.UUID
	NodeAddr map[uuid.UUID]string

	State *State
	sh    StateHolder
	//handlers  map[uuid.UUID]map[uint64]func(string, event.Event) error
}

func New(sh StateHolder) *Orchestrator {
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

	return &Orchestrator{
		Conns:    make(map[string]Conn),
		AddrNode: make(map[string]uuid.UUID),
		NodeAddr: make(map[uuid.UUID]string),
		State:    &state,
		sh:       sh,
	}
}

type Profile struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Links []Link    `json:"links"`
}

type Link struct {
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

func (r *Orchestrator) CreateProfile(name string, links []Link) (Profile, error) {
	if len(links) == 0 {
		return Profile{}, errors.New("no links specified")
	}

	prof := Profile{
		Id:    uuid.New(),
		Name:  name,
		Links: links,
	}

	r.State.Profiles[prof.Id] = prof

	err := r.sh.SetState(*r.State)
	if err != nil {
		return Profile{}, err
	}

	return prof, nil
}

func (r *Orchestrator) EnableProfile(id uuid.UUID) error {
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
	for _, link := range prof.Links {
		if slices.Contains(activeOutputIds, link.OutputId) {
			return errors.New("output already in use")
		}
	}

	r.State.ActiveProfiles = append(r.State.ActiveProfiles, id)

	err := r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	for _, link := range prof.Links {
		srcDev := r.State.Nodes[r.inputNodeId(link.InputId)]
		sinkDev := r.State.Nodes[r.outputNodeId(link.OutputId)]

		// TODO: verify src node is connected

		err = r.write(r.Conns[r.NodeAddr[srcDev.Id]], event.SetInputActive{
			Id: link.InputId,
			Outputs: []event.SetInputActiveOutput{
				{
					OutputId: link.OutputId,
					SinkId:   sinkDev.Id,
					Leds:     sinkDev.Outputs[link.OutputId].Leds,
					Config:   link.Config,
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

func (r *Orchestrator) activeOutputs() []uuid.UUID {
	var outputIds []uuid.UUID

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]
		for _, link := range prof.Links {
			outputIds = append(outputIds, link.OutputId)
		}
	}

	return outputIds
}

func (r *Orchestrator) activeSourceOutputs(id uuid.UUID) []uuid.UUID {
	var outputIds []uuid.UUID

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]
		for _, link := range prof.Links {
			srcId := r.inputNodeId(link.InputId)

			if srcId != id {
				continue
			}

			outputIds = append(outputIds, link.OutputId)
		}
	}

	return outputIds
}

func (r *Orchestrator) activeInputConfigs(id uuid.UUID) []Link {
	var cfgs []Link

	for _, profId := range r.State.ActiveProfiles {
		prof := r.State.Profiles[profId]

		for _, link := range prof.Links {
			if link.InputId != id {
				continue
			}

			cfgs = append(cfgs, link)
		}
	}

	return cfgs
}

func (r *Orchestrator) outputNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, out := range dev.Outputs {
			if out.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Orchestrator) inputNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, in := range dev.Inputs {
			if in.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Orchestrator) sinkNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, in := range dev.Sinks {
			if in.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}
func (r *Orchestrator) sourceNodeId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Nodes {
		for _, in := range dev.Sources {
			if in.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}
