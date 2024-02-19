package state

import (
	"encoding/json"
	"os"
	"sync"

	"ledctl3/pkg/uuid"
	"ledctl3/registry"
)

type Holder struct {
	stateMux sync.Mutex
}

func NewHolder() *Holder {
	return &Holder{}
}

func (s *Holder) SetState(state registry.State) error {
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

func (s *Holder) GetState() (registry.State, error) {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	b, err := os.ReadFile("./registry.json")
	if err != nil {
		return registry.State{}, err
	}

	var state State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return registry.State{}, err
	}

	return ToRegistryState(state), nil
}

type State struct {
	Nodes          map[uuid.UUID]Node             `json:"nodes"`
	Profiles       map[uuid.UUID]registry.Profile `json:"profiles"`
	ActiveProfiles []uuid.UUID                    `json:"activeProfiles"`
}

type Node struct {
	Id      uuid.UUID            `json:"id"`
	Name    string               `json:"name"`
	Inputs  map[uuid.UUID]Input  `json:"inputs"`
	Outputs map[uuid.UUID]Output `json:"outputs"`
	Sources map[uuid.UUID]Source `json:"sources"`
	Sinks   map[uuid.UUID]Sink   `json:"sinks"`
}

type Source struct {
	Id     uuid.UUID       `json:"id"`
	Config json.RawMessage `json:"config"`
}

type Sink struct {
	Id     uuid.UUID       `json:"id"`
	Config json.RawMessage `json:"config"`
}

type Input struct {
	Id       uuid.UUID       `json:"id"`
	DriverId uuid.UUID       `json:"driverId"`
	Schema   json.RawMessage `json:"schema"`
	Config   json.RawMessage `json:"config"`
}

type Output struct {
	Id       uuid.UUID       `json:"id"`
	DriverId uuid.UUID       `json:"driverId"`
	Leds     int             `json:"leds"`
	Schema   json.RawMessage `json:"schema"`
	Config   json.RawMessage `json:"config"`
}

func ToRegistryState(p State) registry.State {
	state := registry.State{
		Nodes:          make(map[uuid.UUID]*registry.Node, len(p.Nodes)),
		Profiles:       p.Profiles,
		ActiveProfiles: p.ActiveProfiles,
	}

	for id, node := range p.Nodes {
		inputs := make(map[uuid.UUID]*registry.Input, len(node.Inputs))
		for _, in := range node.Inputs {
			inputs[in.Id] = &registry.Input{
				Id:       in.Id,
				DriverId: in.DriverId,
				Schema:   in.Schema,
				Config:   in.Config,
			}
		}

		outputs := make(map[uuid.UUID]*registry.Output, len(node.Outputs))
		for _, out := range node.Outputs {
			outputs[out.Id] = &registry.Output{
				Id:       out.Id,
				DriverId: out.DriverId,
				Leds:     out.Leds,
				Schema:   out.Schema,
				Config:   out.Config,
			}
		}

		sources := make(map[uuid.UUID]*registry.Source, len(node.Sources))
		for _, source := range node.Sources {
			sources[source.Id] = &registry.Source{
				Id:     source.Id,
				Config: source.Config,
			}
		}

		sinks := make(map[uuid.UUID]*registry.Sink, len(node.Sinks))
		for _, sink := range node.Sinks {
			sinks[sink.Id] = &registry.Sink{
				Id:     sink.Id,
				Config: sink.Config,
			}
		}

		state.Nodes[id] = &registry.Node{
			Id:      node.Id,
			Name:    node.Name,
			Inputs:  inputs,
			Outputs: outputs,
			Sources: sources,
			Sinks:   sinks,
		}
	}

	return state
}
