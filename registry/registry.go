package registry

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"ledctl3/registry/types"
)

type StateHolder interface {
	SetState(state *State) error
	GetState() (*State, error)
}

type State struct {
	Nodes    map[uuid.UUID]*types.Node    `json:"nodes"`
	Sources  map[uuid.UUID]*types.Source  `json:"sources"`
	Sinks    map[uuid.UUID]*types.Sink    `json:"sinks"`
	Inputs   map[uuid.UUID]*types.Input   `json:"inputs"`
	Outputs  map[uuid.UUID]*types.Output  `json:"outputs"`
	Profiles map[uuid.UUID]*types.Profile `json:"profiles"`
	Links    map[uuid.UUID]*types.Link    `json:"links"`
}

type Registry struct {
	mux sync.Mutex

	Conns map[string]Conn

	AddrNode map[string]uuid.UUID
	NodeAddr map[uuid.UUID]string

	State *State
	sh    StateHolder
}

func New(sh StateHolder) *Registry {
	state, err := sh.GetState()
	if err == nil {
		//fmt.Println("Loaded State", State)
	} else {
		fmt.Println("error reading State", err)
		state = nil
	}

	if state == nil {
		state = &State{
			Nodes:    make(map[uuid.UUID]*types.Node),
			Sources:  make(map[uuid.UUID]*types.Source),
			Sinks:    make(map[uuid.UUID]*types.Sink),
			Inputs:   make(map[uuid.UUID]*types.Input),
			Outputs:  make(map[uuid.UUID]*types.Output),
			Profiles: make(map[uuid.UUID]*types.Profile),
			Links:    make(map[uuid.UUID]*types.Link),
		}
	}

	//fmt.Println("Starting with State", fmt.Sprintf("%#v", State))

	return &Registry{
		Conns:    make(map[string]Conn),
		AddrNode: make(map[string]uuid.UUID),
		NodeAddr: make(map[uuid.UUID]string),
		State:    state,
		sh:       sh,
	}
}
