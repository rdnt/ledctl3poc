package registry

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
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
	Devices        map[uuid.UUID]*Device `json:"devices"`
	Profiles       map[uuid.UUID]Profile `json:"profiles"`
	ActiveProfiles []uuid.UUID           `json:"-"` // TODO: persist active profiles for restarting if registry restarts
}

type Registry struct {
	mux       sync.Mutex
	conns     map[string]uuid.UUID
	connsAddr map[uuid.UUID]string
	write     func(addr string, e event.Event) error
	state     *State
	sh        StateHolder
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
		conns:     make(map[string]uuid.UUID),
		connsAddr: make(map[uuid.UUID]string),
		state:     &state,
		write:     write,
		sh:        sh,
	}
}

func (r *Registry) ProcessEvent(addr string, e event.Event) {
	r.mux.Lock()
	defer r.mux.Unlock()

	//fmt.Println("ProcessEvents")

	switch e := e.(type) {
	case event.Connect:
		r.handleConnect(addr, e)
	case event.Disconnect:
		r.handleDisconnect(addr, e)
	case event.InputAdded:
		r.handleInputConnected(addr, e)
	case event.InputDisconnected:
		r.handleInputDisconnected(addr, e)
	case event.OutputConnected:
		r.handleOutputConnected(addr, e)
	case event.OutputDisconnected:
		r.handleOutputDisconnected(addr, e)
	case event.Data:
		r.handleData(addr, e)
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

func (r *Registry) handleConnect(addr string, e event.Connect) {
	fmt.Printf("%s: recv Connect\n", addr)

	r.conns[addr] = e.Id
	r.connsAddr[e.Id] = addr

	if dev, ok := r.state.Devices[e.Id]; ok {
		dev.Connect()
		r.state.Devices[e.Id] = dev

		fmt.Println("device Connected:", e.Id)

		io := map[uuid.UUID][]uuid.UUID{}
		for _, id := range r.state.ActiveProfiles {
			prof, ok := r.state.Profiles[id]
			if !ok {
				continue
			}

			for inputId, outputIds := range prof.InputOutput {
				io[inputId] = append(io[inputId], outputIds...)
			}
		}

		// TODO: an input can be offline, so this should be done when an input itself
		// connects, not when the source device connects.

		//sourceInputs := map[uuid.UUID][]event.SetSourceActiveInput{}
		//
		//for inputId, outputIds := range io {
		//
		//	sourceId := r.inputDeviceId(inputId)
		//	if sourceId != dev.Id {
		//		continue
		//	}
		//
		//	var outputs []event.SetSourceActiveOutput
		//	for _, outputId := range outputIds {
		//		sinkId := r.outputDeviceId(outputId)
		//		outputs = append(outputs, event.SetSourceActiveOutput{
		//			Id:     outputId,
		//			SinkId: sinkId,
		//			Leds:   r.state.Devices[sinkId].Outputs[outputId].Leds,
		//			Config: nil,
		//		})
		//	}
		//
		//	sourceInputs[sourceId] = append(sourceInputs[sourceId], event.SetSourceActiveInput{
		//		Id:      inputId,
		//		Outputs: outputs,
		//	})
		//}
		//
		//for sourceId, inputs := range sourceInputs {
		//	addr, ok := r.connsAddr[sourceId]
		//	if !ok {
		//		fmt.Println("source device not connected")
		//		continue
		//	}
		//
		//	err := r.send(addr, event.SetSourceActive{
		//		Inputs: inputs,
		//	})
		//	if err != nil {
		//		fmt.Println("error sending event:", err)
		//		continue
		//	}
		//
		//	fmt.Println("sent SetSourceActive to", addr)
		//}

		return
	}

	fmt.Println("unknown Device Connected:", e.Id)

	r.state.Devices[e.Id] = NewDevice(e.Id, true)

	fmt.Println("device added:", e.Id)
	return
}

func (r *Registry) handleDisconnect(addr string, _ event.Disconnect) {
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

	delete(r.conns, addr)
	delete(r.connsAddr, id)
}

func (r *Registry) handleInputConnected(addr string, e event.InputAdded) {
	fmt.Printf("%s: recv InputAdded\n", addr)

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

	var outputIds []uuid.UUID

	for _, id := range r.state.ActiveProfiles {
		prof, ok := r.state.Profiles[id]
		if !ok {
			continue
		}

		for inputId, outIds := range prof.InputOutput {
			if inputId != e.Id {
				continue
			}

			outputIds = append(outputIds, outIds...)
		}
	}

	var outputs []event.SetSourceActiveOutput
	for _, outputId := range outputIds {
		sinkId := r.outputDeviceId(outputId)
		outputs = append(outputs, event.SetSourceActiveOutput{
			Id:     outputId,
			SinkId: sinkId,
			Leds:   r.state.Devices[sinkId].Outputs[outputId].Leds,
			Config: nil,
		})
	}

	err := r.send(addr, event.SetSourceActive{
		Inputs: []event.SetSourceActiveInput{
			{
				Id:      e.Id,
				Outputs: outputs,
			},
		},
	})
	if err != nil {
		fmt.Println("error sending event:", err)
		return
	}

	fmt.Println("sent SetSourceActive to", addr)
}

func (r *Registry) handleInputDisconnected(addr string, e event.InputDisconnected) {
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

	r.state.ActiveProfiles = []uuid.UUID{}
}

func (r *Registry) handleOutputConnected(addr string, e event.OutputConnected) {
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

func (r *Registry) handleOutputDisconnected(addr string, e event.OutputDisconnected) {
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
	Id          uuid.UUID                 `json:"id"`
	Name        string                    `json:"name"`
	InputOutput map[uuid.UUID][]uuid.UUID `json:"io"`
}

//type Profile struct {
//	Id      uuid.UUID       `json:"id"`
//	Name    string          `json:"name"`
//	Sources []ProfileSource `json:"sources"`
//}
//
//type ProfileSource struct {
//	Id     uuid.UUID      `json:"id"`
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

func (r *Registry) AddProfile(prof Profile) (Profile, error) {
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

	if slices.Contains(r.state.ActiveProfiles, id) {
		return errors.New("profile already active")
	}

	activeOutputIds := r.activeOutputs()

	// cannot have multiple outputs active with different inputs.
	// maybe in the future, input data can be combined to the same
	// output with some transformer function, e.g. one input
	// modifying hue/sat and another modifying brightness.
	for _, outputIds := range prof.InputOutput {
		for _, outputId := range outputIds {
			if slices.Contains(activeOutputIds, outputId) {
				return errors.New("output already in use")
			}
		}
	}

	r.state.ActiveProfiles = append(r.state.ActiveProfiles, id)

	err := r.sh.SetState(*r.state)
	if err != nil {
		fmt.Println("error writing state", err)
	}

	sourceInputs := map[uuid.UUID][]event.SetSourceActiveInput{}

	for inputId, outputIds := range prof.InputOutput {

		var outputs []event.SetSourceActiveOutput
		for _, outputId := range outputIds {
			sinkId := r.outputDeviceId(outputId)
			outputs = append(outputs, event.SetSourceActiveOutput{
				Id:     outputId,
				SinkId: sinkId,
				Leds:   r.state.Devices[sinkId].Outputs[outputId].Leds,
				Config: nil,
			})
		}

		sourceId := r.inputDeviceId(inputId)

		sourceInputs[sourceId] = append(sourceInputs[sourceId], event.SetSourceActiveInput{
			Id:      inputId,
			Outputs: outputs,
		})
	}

	for sourceId, inputs := range sourceInputs {
		addr, ok := r.connsAddr[sourceId]
		if !ok {
			fmt.Println("source device not connected")
			continue
		}

		err = r.send(addr, event.SetSourceActive{
			Inputs: inputs,
		})
		if err != nil {
			fmt.Println("error sending event:", err)
			continue
		}

		fmt.Println("sent SetSourceActive to", addr)
	}

	fmt.Println("profile activated:", id)

	return nil
}

func (r *Registry) activeOutputs() []uuid.UUID {
	var outputIds []uuid.UUID

	for _, profId := range r.state.ActiveProfiles {
		prof := r.state.Profiles[profId]
		for _, ids := range prof.InputOutput {
			outputIds = append(outputIds, ids...)
		}
	}

	return outputIds
}

func (r *Registry) outputDeviceId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.state.Devices {
		for _, out := range dev.Outputs {
			if out.Id == id {
				return dev.Id
			}
		}
	}
	return uuid.Nil
}

func (r *Registry) inputDeviceId(id uuid.UUID) uuid.UUID {
	for _, dev := range r.state.Devices {
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

	err := r.send(addr, e)
	if err != nil {
		fmt.Println("error sending event:", err)
		return
	}
}