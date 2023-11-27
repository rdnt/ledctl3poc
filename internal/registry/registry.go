package registry

import (
	"encoding/gob"
	"errors"
	"fmt"
	"slices"
	"sync"

	"ledctl3/event"
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
	Devices        map[uuid.UUID]*Device `json:"devices"`
	Profiles       map[uuid.UUID]Profile `json:"profiles"`
	ActiveProfiles []uuid.UUID           `json:"-"` // TODO: persist active profiles for restarting if registry restarts
}

type Registry struct {
	mux       sync.Mutex
	conns     map[string]uuid.UUID
	connsAddr map[uuid.UUID]string
	write     func(addr string, e event.Event) error
	State     *State
	sh        StateHolder
}

func New(sh StateHolder, write func(addr string, e event.Event) error) *Registry {
	state, err := sh.GetState()
	if err == nil {
		//fmt.Println("Loaded State", State)
	} else {
		fmt.Println("error reading State", err)
		state = State{}
	}

	if state.Devices == nil {
		state.Devices = make(map[uuid.UUID]*Device)
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
		sh:        sh,
	}
}

func (r *Registry) ProcessEvent(addr string, e event.Event) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	//fmt.Println("ProcessEvents")

	var err error
	switch e := e.(type) {
	case event.Connect:
		err = r.handleConnect(addr, e)
	case event.Disconnect:
		err = r.handleDisconnect(addr, e)
	case event.InputConnected:
		err = r.handleInputConnected(addr, e)
	case event.InputDisconnected:
		err = r.handleInputDisconnected(addr, e)
	case event.OutputConnected:
		err = r.handleOutputConnected(addr, e)
	case event.OutputDisconnected:
		err = r.handleOutputDisconnected(addr, e)
	case event.Data:
		r.handleData(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	if err != nil {
		return err
	}

	//fmt.Println("Saving State", fmt.Sprintf("%#v", *r.State))
	err = r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	//fmt.Println("ProcessEvents done")
	return nil
}

func (r *Registry) send(addr string, e any) error {
	_, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	return r.write(addr, e)
}

var ErrDeviceConnected = errors.New("device already connected")

func (r *Registry) handleConnect(addr string, e event.Connect) error {
	fmt.Printf("%s: recv Connect\n", addr)

	if _, ok := r.conns[addr]; ok {
		return ErrDeviceConnected
	}

	r.conns[addr] = e.Id
	r.connsAddr[e.Id] = addr

	if dev, ok := r.State.Devices[e.Id]; ok {
		dev.Connect()
		r.State.Devices[e.Id] = dev

		fmt.Println("device connected:", e.Id)

		return nil
	}

	r.State.Devices[e.Id] = NewDevice(e.Id, true)

	fmt.Println("device added:", e.Id)

	return nil
}

func (r *Registry) handleDisconnect(addr string, _ event.Disconnect) error {
	fmt.Printf("%s: recv Disconnect\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("device already disconnected")
	}

	dev := r.State.Devices[id]

	dev.Disconnect()

	delete(r.conns, addr)
	delete(r.connsAddr, id)

	return nil
}

func (r *Registry) handleInputConnected(addr string, e event.InputConnected) error {
	fmt.Printf("%s: recv InputConnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	dev := r.State.Devices[id]

	dev.ConnectInput(e.Id, e.Schema, e.Config)

	cfgs := r.activeInputConfigs(id)
	if len(cfgs) == 0 {
		return nil
	}

	var evtOutCfgs []event.SetInputActiveOutput
	for _, cfg := range cfgs {
		sink := r.State.Devices[r.outputDeviceId(cfg.OutputId)]

		evtOutCfgs = append(evtOutCfgs, event.SetInputActiveOutput{
			Id:     cfg.OutputId,
			Leds:   sink.Outputs[cfg.OutputId].Leds,
			Config: cfg.Config,
		})
	}

	err := r.send(addr, event.SetInputActive{
		Id:      id,
		Outputs: evtOutCfgs,
	})
	if err != nil {
		return err
	}

	fmt.Println("sent SetInputActive to", addr)

	return nil
}

func (r *Registry) handleInputDisconnected(addr string, e event.InputDisconnected) error {
	fmt.Printf("%s: recv InputDisconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	dev := r.State.Devices[id]

	dev.DisconnectInput(e.Id)

	return nil
}

func (r *Registry) handleOutputConnected(addr string, e event.OutputConnected) error {
	fmt.Printf("%s: recv OutputConnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	dev := r.State.Devices[id]

	dev.ConnectOutput(e.Id, e.Leds, e.Schema, e.Config)

	return nil
}

func (r *Registry) handleOutputDisconnected(addr string, e event.OutputDisconnected) error {
	fmt.Printf("%s: recv OutputDisconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	dev := r.State.Devices[id]

	dev.DisconnectOutput(e.Id)

	return nil
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
//	Id      uuid.UUID       `json:"id"`
//	Name    string          `json:"name"`
//	Sources []ProfileSource `json:"sources"`
//}
//
//type ProfileSource struct {
//	Id     uuid.UUID 	     `json:"id"`
//	Inputs []ProfileInput `json:"inputs"`
//}
//
//type ProfileInput struct {
//	Id    uuid.UUID     `json:"id"`
//	Sinks []ProfileSink `json:"sinks"`
//}
//
//type ProfileSink struct {
//	Id      uuid.UUID       `json:"id"`
//	Outputs []ProfileOutput `json:"outputs"`
//}
//
//type ProfileOutput struct {
//	Id            uuid.UUID `json:"id"`
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
		srcDev := r.State.Devices[r.inputDeviceId(io.InputId)]
		sinkDev := r.State.Devices[r.outputDeviceId(io.OutputId)]

		err = r.send(r.connsAddr[srcDev.Id], event.SetInputActive{
			Id: io.InputId,
			Outputs: []event.SetInputActiveOutput{
				{
					Id:     io.OutputId,
					Leds:   sinkDev.Outputs[io.OutputId].Leds,
					Config: io.Config,
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

func (r *Registry) outputDeviceId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Devices {
		for _, out := range dev.Outputs {
			if out.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Registry) inputDeviceId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.State.Devices {
		for _, in := range dev.Inputs {
			if in.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Registry) handleData(_ string, e event.Data) {
	addr, ok := r.connsAddr[e.SinkId]
	if !ok {
		fmt.Println("unknown sink device:", e.SinkId)
		return
	}

	fmt.Print(".")
	err := r.send(addr, e)
	if err != nil {
		fmt.Println("error sending event:", err)
		return
	}
}
