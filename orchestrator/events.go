package orchestrator

import (
	"errors"
	"fmt"
	"slices"
	"time"

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

	//fmt.Println("HandleConnection")

	var err error
	switch e := e.(type) {
	case ConnectedEvent:
		err = r.handleConnected(c)
	case DisconnectedEvent:
		err = r.handleDisconnected(c)
	case event.NodeConnected:
		err = r.handleNodeConnected(c, e)
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

	//fmt.Println("Saving State", fmt.Sprintf("%#v", *r.State))
	err = r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	//fmt.Println("HandleConnection done")
	return nil
}

// TODO: remove
func (r *Orchestrator) send(addr string, e event.Event) error {
	_, ok := r.AddrNode[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	return nil
	//return r.write(addr, e)
}

// TODO: remove
func (r *Orchestrator) req(addr string, e event.Event) error {
	_, ok := r.AddrNode[addr]
	if !ok {
		return errors.New("node disconnected")
	}

	return nil
	//return r.request(addr, e)
}

func (r *Orchestrator) handleNodeConnected(c Conn, e event.NodeConnected) error {
	fmt.Printf("%s: recv NodeConnected\n", c.Addr())

	if _, ok := r.AddrNode[c.Addr()]; ok {
		return errors.New("node already connected")
	}

	if e.Id == uuid.Nil {
		return errors.New("invalid node id")
	}

	r.AddrNode[c.Addr()] = e.Id
	r.NodeAddr[e.Id] = c.Addr()

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

type Conn interface {
	Addr() string
	Latency() time.Duration
	Write(b []byte) error
}

func (r *Orchestrator) handleConnected(c Conn) error {
	fmt.Printf("%s: recv Connected\n", c.Addr())

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

	srcId, ok := r.AddrNode[c.Addr()]
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

	err := r.send(c.Addr(), event.SetInputActive{
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

	id, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
	}

	dev := r.State.Nodes[id]

	dev.DisconnectInput(e.Id)

	return nil
}

func (r *Orchestrator) handleOutputConnected(c Conn, e event.OutputConnected) error {
	fmt.Printf("%s: recv OutputConnected\n", c.Addr())

	id, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
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
	srcId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
	}

	sinkDev := r.State.Nodes[e.SinkId]
	if sinkDev == nil {
		return errors.New("unknown sink node")
	}

	sinkAddr, ok := r.NodeAddr[e.SinkId]
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
