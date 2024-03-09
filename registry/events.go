package registry

import (
	"errors"
	"fmt"
	"slices"

	"ledctl3/node/event"
	"ledctl3/pkg/uuid"
)

type ConnectedEvent struct{}

type DisconnectedEvent struct{}

func (r *Registry) ProcessEvent(addr string, e event.Event) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	//fmt.Println("HandleConnection")

	var err error
	switch e := e.(type) {
	//case ConnectedEvent:
	//	r.HandleConnected(addr)
	//case DisconnectedEvent:
	//	r.HandleDisconnected(addr)
	case event.NodeConnected:
		err = r.handleNodeConnected(addr, e)
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
	err = r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	//fmt.Println("HandleConnection done")
	return nil
}

func (r *Registry) send(addr string, e event.Event) error {
	_, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	return r.write(addr, e)
}

func (r *Registry) req(addr string, e event.Event) error {
	_, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	return r.request(addr, e)
}

func (r *Registry) handleNodeConnected(addr string, e event.NodeConnected) error {
	fmt.Printf("%s: recv NodeConnected\n", addr)

	if _, ok := r.conns[addr]; ok {
		return errors.New("node already connected")
	}

	r.conns[addr] = e.Id
	r.connsAddr[e.Id] = addr

	if dev, ok := r.State.Nodes[e.Id]; ok {
		dev.Connect()

		dev.Sources = make(map[uuid.UUID]*Source)
		for _, d := range e.Sources {
			dev.Sources[d.Id] = &Source{d.Id, d.Config, true}
		}

		dev.Sinks = make(map[uuid.UUID]*Sink)
		for _, d := range e.Sinks {
			dev.Sinks[d.Id] = &Sink{d.Id, d.Config, true}
		}

		r.State.Nodes[e.Id] = dev

		fmt.Println("node connected:", e.Id)

		return nil
	}

	sources := make(map[uuid.UUID]*Source)
	for _, d := range e.Sources {
		sources[d.Id] = &Source{d.Id, d.Config, true}
	}

	sinks := make(map[uuid.UUID]*Sink)
	for _, d := range e.Sinks {
		sinks[d.Id] = &Sink{d.Id, d.Config, true}
	}

	r.State.Nodes[e.Id] = NewNode(e.Id, true, sources, sinks)

	fmt.Println("node added:", e.Id)

	return nil
}

func (r *Registry) HandleConnected(addr string) {
	fmt.Printf("%s: recv Connected\n", addr)
}

func (r *Registry) HandleDisconnected(addr string) {
	fmt.Printf("%s: recv Disconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return
	}

	dev := r.State.Nodes[id]

	dev.Disconnect()

	delete(r.conns, addr)
	delete(r.connsAddr, id)
}

func (r *Registry) handleInputConnected(addr string, e event.InputConnected) error {
	fmt.Printf("%s: recv InputConnected\n", addr)

	srcId, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[srcId]

	dev.ConnectInput(e.Id, e.DriverId, e.Schema, e.Config)

	cfgs := r.activeInputConfigs(e.Id)
	if len(cfgs) == 0 {
		return nil
	}

	var evtOutCfgs []event.SetInputActiveOutput
	for _, cfg := range cfgs {
		dev := r.State.Nodes[r.outputNodeId(cfg.OutputId)]

		evtOutCfgs = append(evtOutCfgs, event.SetInputActiveOutput{
			OutputId: cfg.OutputId,
			SinkId:   dev.Outputs[cfg.OutputId].DriverId,
			NodeId:   dev.Id,
			Leds:     dev.Outputs[cfg.OutputId].Leds,
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
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[id]

	dev.DisconnectInput(e.Id)

	return nil
}

func (r *Registry) handleOutputConnected(addr string, e event.OutputConnected) error {
	fmt.Printf("%s: recv OutputConnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[id]

	dev.ConnectOutput(e.Id, e.DriverId, e.Leds, e.Schema, e.Config)

	return nil
}

func (r *Registry) handleOutputDisconnected(addr string, e event.OutputDisconnected) error {
	fmt.Printf("%s: recv OutputDisconnected\n", addr)

	id, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[id]

	dev.DisconnectOutput(e.Id)

	return nil
}

func (r *Registry) handleData(addr string, e event.Data) error {
	srcId, ok := r.conns[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	sinkDev := r.State.Nodes[e.SinkId]
	if sinkDev == nil {
		return errors.New("unknown sink node")
	}

	sinkAddr, ok := r.connsAddr[e.SinkId]
	if !ok {
		return errors.New("sink node disconnected")
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

	//fmt.Println("latency", e.Latency)

	go func() {
		err := r.send(sinkAddr, e)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
