package state

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/uuid"

	"ledctl3/registry"
	"ledctl3/registry/types"
)

type Holder struct {
	stateMux sync.Mutex
}

func NewHolder() *Holder {
	return &Holder{}
}

func (s *Holder) SetState(state *registry.State) error {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("./registry.json", b, 0644)
}

func (s *Holder) Stop() error {
	st, err := s.GetState()
	if err != nil {
		return err
	}

	return s.SetState(st)
}

func (s *Holder) GetState() (*registry.State, error) {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	b, err := os.ReadFile("./registry.json")
	if err != nil {
		return nil, err
	}

	var state *registry.State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	return ToRegistryState(state), nil
}

func ToRegistryState(p *registry.State) *registry.State {
	if p.Nodes == nil {
		p.Nodes = make(map[uuid.UUID]*types.Node)
	}
	for id := range p.Nodes {
		p.Nodes[id].Connected = false
	}

	if p.Sources == nil {
		p.Sources = make(map[uuid.UUID]*types.Source)
	}
	for id := range p.Sources {
		p.Sources[id].Connected = false
	}

	if p.Sinks == nil {
		p.Sinks = make(map[uuid.UUID]*types.Sink)
	}
	for id := range p.Sinks {
		p.Sinks[id].Connected = false
	}

	if p.Inputs == nil {
		p.Inputs = make(map[uuid.UUID]*types.Input)
	}
	for id := range p.Inputs {
		p.Inputs[id].Connected = false
	}

	if p.Outputs == nil {
		p.Outputs = make(map[uuid.UUID]*types.Output)
	}
	for id := range p.Outputs {
		p.Outputs[id].Connected = false
	}

	if p.Profiles == nil {
		p.Profiles = make(map[uuid.UUID]*types.Profile)
	}
	if p.Links == nil {
		p.Links = make(map[uuid.UUID]*types.Link)
	}

	return p
}
