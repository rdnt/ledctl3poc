package registry

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"ledctl3/event"
	"ledctl3/pkg/uuid"
)

type sh struct {
}

func init() {
	gob.Register(State{})
}

func (s sh) SetState(state State) error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("registry.json", b, 0644)
}

func (s sh) GetState() (State, error) {
	b, err := os.ReadFile("registry.json")
	if err != nil {
		return State{}, err
	}

	var state State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return State{}, err
	}

	return state, nil
}

type StateHolder interface {
	SetState(state State) error
	GetState() (State, error)
}

type State struct {
	Devices  map[uuid.UUID]*Device `json:"devices"`
	Profiles map[uuid.UUID]Profile `json:"profiles"`
}

type Registry struct {
	mux   sync.Mutex
	conns map[string]uuid.UUID
	write func(addr string, e event.Event) error
	state *State
	sh    StateHolder
}

func New(write func(addr string, e event.Event) error) *Registry {
	sh := sh{}
	state, err := sh.GetState()
	if err == nil {
		//fmt.Println("Loaded state", state)
	} else {
		fmt.Println("error reading state", err)
		state = State{
			Devices:  make(map[uuid.UUID]*Device),
			Profiles: make(map[uuid.UUID]Profile),
		}
	}

	fmt.Println("Starting with state", fmt.Sprintf("%#v", state))

	return &Registry{
		conns: make(map[string]uuid.UUID),
		state: &state,
		write: write,
		sh:    sh,
	}
}

func (r *Registry) ProcessEvent(addr string, e event.Event) {
	r.mux.Lock()
	defer r.mux.Unlock()

	//fmt.Println("ProcessEvents")

	switch e := e.(type) {
	case event.Connect:
		r.handleConnectEvent(addr, e)
	case event.Disconnect:
		r.handleDisconnectEvent(addr, e)
	case event.InputConnected:
		r.handleInputConnectedEvent(addr, e)
	case event.InputDisconnected:
		r.handleInputDisconnectedEvent(addr, e)
	case event.OutputConnected:
		r.handleOutputConnectedEvent(addr, e)
	case event.OutputDisconnected:
		r.handleOutputDisconnectedEvent(addr, e)
	default:
		fmt.Println("unknown event", e)
	}

	//fmt.Println("Saving state", fmt.Sprintf("%#v", *r.state))
	err := r.sh.SetState(*r.state)
	if err != nil {
		fmt.Println("error writing state", err)
	}

	//fmt.Println("ProcessEvents done")
}

func (r *Registry) send(addr string, e any) error {
	_, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	return r.write(addr, e)
}

func (r *Registry) handleConnectEvent(addr string, e event.Connect) {
	fmt.Printf("%s: recv Connect\n", addr)

	r.conns[addr] = e.Id

	if dev, ok := r.state.Devices[e.Id]; ok {
		dev.Connect()
		r.state.Devices[e.Id] = dev
		fmt.Println("device Connected:", e.Id)
		return
	}

	fmt.Println("unknown Device Connected:", e.Id)

	r.state.Devices[e.Id] = NewDevice(e.Id, true)

	fmt.Println("device added:", e.Id)
	return
}

func (r *Registry) handleDisconnectEvent(addr string, _ event.Disconnect) {
	fmt.Printf("%s: recv Disconnect\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		fmt.Println("unknown conn disconnected", addr)
		return
	}

	dev, ok := r.state.Devices[id]
	if !ok {
		fmt.Println("unknown Device:", id)
		return
	}

	dev.Disconnect()
	r.state.Devices[id] = dev
}

func (r *Registry) handleInputConnectedEvent(addr string, e event.InputConnected) {
	fmt.Printf("%s: recv InputConnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		fmt.Println("unknown connection:", addr)
		return
	}

	dev, ok := r.state.Devices[id]
	if !ok {
		fmt.Println("unknown Device:", id)
		return
	}

	dev.ConnectInput(e.Id, string(e.Type))
	r.state.Devices[id] = dev
}

func (r *Registry) handleInputDisconnectedEvent(addr string, e event.InputDisconnected) {
	fmt.Printf("%s: recv InputDisconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		fmt.Println("unknown connection:", addr)
		return
	}

	dev, ok := r.state.Devices[id]
	if !ok {
		fmt.Println("unknown device:", id)
		return
	}

	dev.DisconnectInput(e.Id)
	r.state.Devices[id] = dev
}

func (r *Registry) handleOutputConnectedEvent(addr string, e event.OutputConnected) {
	fmt.Printf("%s: recv OutputConnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		fmt.Println("unknown connection:", addr)
		return
	}

	dev, ok := r.state.Devices[id]
	if !ok {
		fmt.Println("unknown Device:", id)
		return
	}

	dev.ConnectOutput(e.Id, e.Leds)
	r.state.Devices[id] = dev
}

func (r *Registry) handleOutputDisconnectedEvent(addr string, e event.OutputDisconnected) {
	fmt.Printf("%s: recv OutputDisconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		fmt.Println("unknown connection:", addr)
		return
	}

	dev, ok := r.state.Devices[id]
	if !ok {
		fmt.Println("unknown Device:", id)
		return
	}

	dev.DisconnectOutput(e.Id)
	r.state.Devices[id] = dev
}

type Profile struct {
	Id      uuid.UUID       `json:"id"`
	Name    string          `json:"name"`
	Sources []ProfileSource `json:"sources"`
}

type ProfileSource struct {
	Id     uuid.UUID      `json:"id"`
	Inputs []ProfileInput `json:"inputs"`
}

type ProfileInput struct {
	Id    uuid.UUID     `json:"id"`
	Sinks []ProfileSink `json:"sinks"`
}

type ProfileSink struct {
	Id      uuid.UUID       `json:"id"`
	Outputs []ProfileOutput `json:"outputs"`
}

type ProfileOutput struct {
	Id            uuid.UUID `json:"id"`
	InputConfigId uuid.UUID `json:"input_config_id"`
}

func (r *Registry) AddProfile(name string, sources []ProfileSource) (Profile, error) {
	prof := Profile{
		Id:      uuid.New(),
		Name:    name,
		Sources: sources,
	}

	if r.state.Profiles == nil {
		r.state.Profiles = make(map[uuid.UUID]Profile)
	}

	r.state.Profiles[prof.Id] = prof

	err := r.sh.SetState(*r.state)
	if err != nil {
		fmt.Println("error writing state", err)
	}

	return prof, nil
}

func (r *Registry) SelectProfile(id uuid.UUID) error {
	prof, ok := r.state.Profiles[id]
	if !ok {
		return errors.New("profile not found")
	}

	// verify all devices and IO are connected
	for _, source := range prof.Sources {
		dev, ok := r.state.Devices[source.Id]
		if !ok {
			return errors.New("source device not found")
		}

		if !dev.Connected {
			return errors.New("source device not connected")
		}

		for _, input := range source.Inputs {
			if !dev.Inputs[input.Id].Connected {
				return errors.New("input not connected")
			}

			for _, sink := range input.Sinks {
				dev, ok := r.state.Devices[sink.Id]
				if !ok {
					return errors.New("sink device not found")
				}

				if !dev.Connected {
					return errors.New("sink device not connected")
				}

				for _, output := range sink.Outputs {
					if !dev.Outputs[output.Id].Connected {
						return errors.New("output not connected")
					}
				}
			}
		}
	}

	var startErrs []error
	for _, source := range prof.Sources {
		conn, ok := r.conns[source.Id.String()]
		if !ok {
			startErrs = append(startErrs, errors.New("source device not connected"))
			continue
		}

		err := r.send(conn.String(), event.SetSourceActive{})
		if err != nil {
			startErrs = append(startErrs, err)
			continue
		}
	}

	if len(startErrs) > 0 {
		return fmt.Errorf("failed to start stream: %v", startErrs)
	}

	fmt.Println("All connected!")

	return nil
}
