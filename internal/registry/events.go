package registry

import (
	"errors"
	"fmt"
	"slices"

	"ledctl3/event"
)

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
		err = r.handleData(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	if err != nil {
		return err
	}

	//fmt.Println("Saving State", fmt.Sprintf("%#v", *r.State))
	//err = r.sh.SetState(*r.State)
	//if err != nil {
	//	fmt.Println("error writing State", err)
	//}

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

func (r *Registry) handleConnect(addr string, e event.Connect) error {
	fmt.Printf("%s: recv Connect\n", addr)

	if _, ok := r.conns[addr]; ok {
		return errors.New("device already connected")
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

	srcId, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	dev := r.State.Devices[srcId]

	dev.ConnectInput(e.Id, e.Schema, e.Config)

	cfgs := r.activeInputConfigs(e.Id)
	if len(cfgs) == 0 {
		return nil
	}

	var evtOutCfgs []event.SetInputActiveOutput
	for _, cfg := range cfgs {
		sink := r.State.Devices[r.outputDeviceId(cfg.OutputId)]

		evtOutCfgs = append(evtOutCfgs, event.SetInputActiveOutput{
			OutputId: cfg.OutputId,
			SinkId:   sink.Id,
			Leds:     sink.Outputs[cfg.OutputId].Leds,
			Config:   cfg.Config,
		})
	}

	err := r.send(addr, event.SetInputActive{
		Id:      e.Id,
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

func (r *Registry) handleData(addr string, e event.Data) error {
	srcId, ok := r.conns[addr]
	if !ok {
		return errors.New("device disconnected")
	}

	sinkDev := r.State.Devices[e.SinkId]
	if sinkDev == nil {
		return errors.New("unknown sink device")
	}

	sinkAddr, ok := r.connsAddr[e.SinkId]
	if !ok {
		return errors.New("sink device disconnected")
	}

	srcOutputs := r.activeSourceOutputs(srcId)

	var valid bool
	for _, out := range e.Outputs {
		if slices.Contains(srcOutputs, out.OutputId) {
			valid = true
			break
		}
	}

	if !valid {
		return errors.New("invalid output")
	}

	fmt.Println("latency", e.Latency)

	go func() {
		err := r.send(sinkAddr, e)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
