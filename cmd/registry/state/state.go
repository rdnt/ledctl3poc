package state

import (
	"encoding/json"
	"os"
	"sync"

	"ledctl3/pkg/uuid"
	registry2 "ledctl3/registry"
)

type Holder struct {
	stateMux sync.Mutex
}

func NewHolder() *Holder {
	return &Holder{}
}

func (s *Holder) SetState(state registry2.State) error {
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

func (s *Holder) GetState() (registry2.State, error) {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	b, err := os.ReadFile("./registry.json")
	if err != nil {
		return registry2.State{}, err
	}

	var state State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return registry2.State{}, err
	}

	return ToRegistryState(state), nil
}

type State struct {
	Nodes          map[uuid.UUID]Node              `json:"nodes"`
	Profiles       map[uuid.UUID]registry2.Profile `json:"profiles"`
	ActiveProfiles []uuid.UUID                     `json:"activeProfiles"`
}

type Node struct {
	Id      uuid.UUID            `json:"id"`
	Name    string               `json:"name"`
	Inputs  map[uuid.UUID]Input  `json:"inputs"`
	Outputs map[uuid.UUID]Output `json:"outputs"`
	Drivers map[uuid.UUID]Driver `json:"drivers"`
}

type Driver struct {
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

func ToState(s registry2.State) State {
	state := State{
		Nodes:          make(map[uuid.UUID]Node, len(s.Nodes)),
		Profiles:       s.Profiles,
		ActiveProfiles: s.ActiveProfiles,
	}

	for id, node := range s.Nodes {
		inputs := make(map[uuid.UUID]Input, len(node.Inputs))
		for _, in := range node.Inputs {
			inputs[in.Id] = Input{
				Id:       in.Id,
				DriverId: in.DriverId,
				Schema:   in.Schema,
				Config:   in.Config,
			}
		}

		outputs := make(map[uuid.UUID]Output, len(node.Outputs))
		for _, out := range node.Outputs {
			outputs[out.Id] = Output{
				Id:       out.Id,
				DriverId: out.DriverId,
				Leds:     out.Leds,
				Schema:   out.Schema,
				Config:   out.Config,
			}
		}

		drivers := make(map[uuid.UUID]Driver, len(node.Drivers))
		for _, driver := range node.Drivers {
			drivers[driver.Id] = Driver{
				Id:     driver.Id,
				Config: driver.Config,
			}
		}

		state.Nodes[id] = Node{
			Id:      node.Id,
			Name:    node.Name,
			Inputs:  inputs,
			Outputs: outputs,
			Drivers: drivers,
		}
	}

	return state
}

func ToRegistryState(p State) registry2.State {
	state := registry2.State{
		Nodes:          make(map[uuid.UUID]*registry2.Node, len(p.Nodes)),
		Profiles:       p.Profiles,
		ActiveProfiles: p.ActiveProfiles,
	}

	for id, node := range p.Nodes {
		inputs := make(map[uuid.UUID]*registry2.Input, len(node.Inputs))
		for _, in := range node.Inputs {
			inputs[in.Id] = &registry2.Input{
				Id:       in.Id,
				DriverId: in.DriverId,
				Schema:   in.Schema,
				Config:   in.Config,
			}
		}

		outputs := make(map[uuid.UUID]*registry2.Output, len(node.Outputs))
		for _, out := range node.Outputs {
			outputs[out.Id] = &registry2.Output{
				Id:       out.Id,
				DriverId: out.DriverId,
				Leds:     out.Leds,
				Schema:   out.Schema,
				Config:   out.Config,
			}
		}

		drivers := make(map[uuid.UUID]*registry2.Driver, len(node.Drivers))
		for _, driver := range node.Drivers {
			drivers[driver.Id] = &registry2.Driver{
				Id:     driver.Id,
				Config: driver.Config,
			}
		}

		state.Nodes[id] = &registry2.Node{
			Id:      node.Id,
			Name:    node.Name,
			Inputs:  inputs,
			Outputs: outputs,
			Drivers: drivers,
		}
	}

	return state
}
