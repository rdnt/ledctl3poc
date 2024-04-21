package orchestrator

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"ledctl3/node/event"
)

type ConnectedEvent struct{}

func (e ConnectedEvent) Type() string {
	return "Connected"
}

type DisconnectedEvent struct{}

func (e DisconnectedEvent) Type() string {
	return "Disconnected"
}

func (r *Orchestrator) ProcessEvent(c Conn, e event.Event) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	// fmt.Println("HandleConnection")

	var err error
	switch e := e.(type) {
	case ConnectedEvent:
		err = r.handleConnected(c)
	case DisconnectedEvent:
		err = r.handleDisconnected(c)
	case event.Authenticate:
		err = r.handleAuthenticate(c, e)
	case event.Capabilities:
		err = r.handleCapabilities(c, e)
	case event.InputConnected:
		err = r.handleInputConnected(c, e)
	case event.InputDisconnected:
		err = r.handleInputDisconnected(c, e)
	case event.OutputConnected:
		err = r.handleOutputConnected(c, e)
	case event.OutputDisconnected:
		err = r.handleOutputDisconnected(c, e)
	case event.Data:
		err = r.handleData(c, e)

	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	if err != nil {
		return err
	}

	// fmt.Println("Saving State", fmt.Sprintf("%#v", *r.State))
	err = r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	// fmt.Println("HandleConnection done")
	return nil
}

func (r *Orchestrator) write(c Conn, v event.Event) error {
	codec := event.JSONCodec{}

	b, err := codec.MarshalBinary(v)
	if err != nil {
		return err
	}

	return c.Write(b)
}

func (r *Orchestrator) createNode(id uuid.UUID) error {
	r.State.Nodes[id] = NewNode(id, true, nil, nil)
	return nil
}

func (r *Orchestrator) handleAuthenticate(c Conn, e event.Authenticate) error {
	fmt.Printf("%s: recv Authenticate\n", c.Addr())

	if e.Id == uuid.Nil {
		return errors.New("authentication failed")
	}

	if _, ok := r.AddrNode[c.Addr()]; ok {
		return nil
	}

	r.AddrNode[c.Addr()] = e.Id
	r.NodeAddr[e.Id] = c.Addr()

	if dev, ok := r.State.Nodes[e.Id]; ok {
		dev.Connect()
		fmt.Println("node connected:", e.Id)

		return nil
	}

	r.State.Nodes[e.Id] = NewNode(e.Id, true, nil, nil)
	fmt.Println("node added:", e.Id)

	return nil
}

func (r *Orchestrator) handleCapabilities(c Conn, e event.Capabilities) error {
	fmt.Printf("%s: recv Capabilities\n", c.Addr())

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	dev, ok := r.State.Nodes[nodeId]
	if !ok {
		return errors.New("node not found")
	}

	// TODO: handle capabilities config updating without removing existing capabilities
	// that are no longer listed

	dev.Sources = make(map[uuid.UUID]*Source)
	for _, d := range e.Sources {
		dev.Sources[d.Id] = &Source{d.Id, d.Config, true}
	}

	dev.Sinks = make(map[uuid.UUID]*Sink)
	for _, d := range e.Sinks {
		dev.Sinks[d.Id] = &Sink{d.Id, d.Config, true}
	}

	fmt.Println("node capabilities updated", nodeId)

	return nil
}

type Conn interface {
	Addr() string
	Write(b []byte) error
	Read() ([]byte, error)
	Ping() error
}

func (r *Orchestrator) handleConnected(c Conn) error {
	fmt.Printf("%s: recv Capabilities\n", c.Addr())

	if _, ok := r.Conns[c.Addr()]; ok {
		return errors.New("client already connected")
	}

	r.Conns[c.Addr()] = c
	return nil
}

func (r *Orchestrator) handleDisconnected(c Conn) error {
	fmt.Printf("%s: recv Disconnected\n", c.Addr())

	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("client already disconnected")
	}

	delete(r.Conns, c.Addr())

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return nil
	}

	node, ok := r.State.Nodes[nodeId]
	if !ok {
		return errors.New("node not found")
	}

	node.Disconnect()

	delete(r.AddrNode, c.Addr())
	delete(r.NodeAddr, nodeId)

	return nil
}

func (r *Orchestrator) handleInputConnected(c Conn, e event.InputConnected) error {
	fmt.Printf("%s: recv InputConnected\n", c.Addr())

	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	srcId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
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

		// TODO: validate output exists

		evtOutCfgs = append(evtOutCfgs, event.SetInputActiveOutput{
			OutputId: cfg.OutputId,
			SinkId:   dev.Outputs[cfg.OutputId].DriverId,
			Leds:     dev.Outputs[cfg.OutputId].Leds,
			Config:   cfg.Config,
		})
	}

	err := r.write(c, event.SetInputActive{
		Id:      e.Id,
		Outputs: evtOutCfgs,
	})
	if err != nil {
		return err
	}

	fmt.Println("sent SetInputActive to", c.Addr())

	return nil
}

func (r *Orchestrator) handleInputDisconnected(c Conn, e event.InputDisconnected) error {
	fmt.Printf("%s: recv InputDisconnected\n", c.Addr())

	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	id, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	dev := r.State.Nodes[id]

	dev.DisconnectInput(e.Id)

	return nil
}

func (r *Orchestrator) handleOutputConnected(c Conn, e event.OutputConnected) error {
	fmt.Printf("%s: recv OutputConnected\n", c.Addr())

	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	id, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	dev := r.State.Nodes[id]

	dev.ConnectOutput(e.Id, e.DriverId, e.Leds, e.Schema, e.Config)

	return nil
}

func (r *Orchestrator) handleOutputDisconnected(c Conn, e event.OutputDisconnected) error {
	fmt.Printf("%s: recv OutputDisconnected\n", c.Addr())

	id, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[id]

	dev.DisconnectOutput(e.Id)

	return nil
}

func (r *Orchestrator) handleData(c Conn, e event.Data) error {
	//fmt.Printf("%s: recv Data\n", c.Addr())

	srcId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
	}

	nodeId := r.sinkNodeId(e.SinkId)

	sinkDev := r.State.Nodes[nodeId]
	if sinkDev == nil {
		return errors.New("unknown sink node")
	}

	sinkAddr, ok := r.NodeAddr[nodeId]
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

	// fmt.Println("latency", e.Latency)

	go func() {
		// TODO verify sink connected
		err := r.write(r.Conns[sinkAddr], e)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
