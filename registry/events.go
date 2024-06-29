package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/google/uuid"

	"ledctl3/node/event"
	"ledctl3/registry/types"
)

type ConnectedEvent struct{}

func (e ConnectedEvent) Type() string {
	return "Connected"
}

type DisconnectedEvent struct{}

func (e DisconnectedEvent) Type() string {
	return "Disconnected"
}

func (r *Registry) ProcessEvent(c Conn, e event.Event) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	if _, ok := e.(event.Data); !ok {
		fmt.Printf("%s: recv %s\n", c.Addr(), reflect.TypeOf(e).Name())
	}

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
	err = r.sh.SetState(r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	// fmt.Println("HandleConnection done")
	return nil
}

func (r *Registry) handleAuthenticate(c Conn, e event.Authenticate) error {
	if e.Id == uuid.Nil {
		return errors.New("authentication failed")
	}

	if _, ok := r.AddrNode[c.Addr()]; ok {
		return nil
	}

	r.AddrNode[c.Addr()] = e.Id
	r.NodeAddr[e.Id] = c.Addr()

	node, ok := r.State.Nodes[e.Id]
	if !ok {
		node = types.NewNode(e.Id)
		r.State.Nodes[e.Id] = node
		fmt.Println("node added:", e.Id)
	}

	node.Connected = true
	node.Address = c.Addr()
	fmt.Println("node connected:", e.Id)

	return nil
}

func (r *Registry) handleCapabilities(c Conn, e event.Capabilities) error {
	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	node, ok := r.State.Nodes[nodeId]
	if !ok {
		return errors.New(fmt.Sprint("node not found ", nodeId))
	}

	srcCfg := map[uuid.UUID]json.RawMessage{}
	sinkCfg := map[uuid.UUID]json.RawMessage{}

	for _, es := range e.Sources {
		source, ok := r.State.Sources[es.Id]
		if !ok {
			source = types.NewSource(es.Id, node.Id)
			r.State.Sources[source.Id] = source
			node.SourceIds = append(node.SourceIds, source.Id)
		}

		source.Schema = es.Schema
		srcCfg[source.Id] = source.Config
	}

	for _, es := range e.Sinks {
		sink, ok := r.State.Sinks[es.Id]
		if !ok {
			sink = types.NewSink(es.Id, node.Id)
			r.State.Sinks[sink.Id] = sink
			node.SinkIds = append(node.SinkIds, sink.Id)
		}

		sink.Schema = es.Schema
		sinkCfg[sink.Id] = sink.Config
	}

	//for sourceId, cfg := range srcCfg {
	//	r.sendSourceConfig(nodeId, sourceId, cfg)
	//}
	//
	//for sinkId, cfg := range sinkCfg {
	//	r.sendSinkConfig(nodeId, sinkId, cfg)
	//}

	fmt.Println("node capabilities updated", nodeId)

	return nil
}

//func (r *Registry) sendSourceConfig(nodeId uuid.UUID, sourceId uuid.UUID, cfg json.RawMessage) {
//	err := r.writeNode(nodeId, event.SetSourceConfig{
//		SourceId: sourceId,
//		Config:   cfg,
//	})
//	if err != nil {
//		fmt.Println("error sending event:", err)
//		return
//	}
//}
//
//func (r *Registry) sendSinkConfig(nodeId uuid.UUID, sourceId uuid.UUID, cfg json.RawMessage) {
//	err := r.writeNode(nodeId, event.SetSinkConfig{
//		SinkId: sourceId,
//		Config: cfg,
//	})
//	if err != nil {
//		fmt.Println("error sending event:", err)
//		return
//	}
//}

type Conn interface {
	Addr() string
	Write(b []byte) error
	Read() ([]byte, error)
	Ping() error
}

func (r *Registry) handleConnected(c Conn) error {
	if _, ok := r.Conns[c.Addr()]; ok {
		return errors.New("client already connected")
	}

	r.Conns[c.Addr()] = c
	return nil
}

func (r *Registry) handleDisconnected(c Conn) error {
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
		return errors.New(fmt.Sprint("node not found ", nodeId))
	}

	node.Connected = false
	node.Address = ""

	delete(r.AddrNode, c.Addr())
	delete(r.NodeAddr, nodeId)

	return nil
}

func (r *Registry) handleInputConnected(c Conn, e event.InputConnected) error {
	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	node := r.State.Nodes[nodeId]

	source, ok := r.State.Sources[e.SourceId]
	if !ok {
		return errors.New("unknown source")
	}

	input, ok := r.State.Inputs[e.Id]
	if !ok {
		input = types.NewInput(e.Id, node.Id, source.Id)
		r.State.Inputs[input.Id] = input
		source.InputIds = append(source.InputIds, input.Id)
	}

	input.Connected = true
	input.Schema = e.Schema

	r.configureInput(nodeId, input.SourceId, input.Id, input.Config)

	return nil
}

func (r *Registry) configureInput(nodeId, sourceId, inputId uuid.UUID, cfg json.RawMessage) {
	err := r.writeNode(nodeId, event.ConfigureInput{
		SourceId: sourceId,
		InputId:  inputId,
		Config:   cfg,
	})
	if err != nil {
		fmt.Println("error sending event:", err)
		return
	}
}

func (r *Registry) handleInputDisconnected(c Conn, e event.InputDisconnected) error {
	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	node := r.State.Nodes[nodeId]

	source, ok := r.State.Sources[e.SourceId]
	if !ok {
		return errors.New("unknown source")
	}

	if source.NodeId != node.Id {
		return errors.New("source not found")
	}

	input, ok := r.State.Inputs[e.Id]
	if !ok {
		return errors.New("unknown input")
	}

	if input.SourceId != e.SourceId {
		return errors.New("input not found")
	}

	input.Connected = false

	return nil
}

func (r *Registry) handleOutputConnected(c Conn, e event.OutputConnected) error {
	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	node := r.State.Nodes[nodeId]

	sink, ok := r.State.Sinks[e.SinkId]
	if !ok {
		return errors.New("unknown sink")
	}

	output, ok := r.State.Outputs[e.Id]
	if !ok {
		output = types.NewOutput(e.Id, node.Id, sink.Id, e.Leds)
		r.State.Outputs[output.Id] = output
		sink.OutputIds = append(sink.OutputIds, output.Id)
	}

	output.Connected = true
	output.Leds = e.Leds
	output.Schema = e.Schema

	r.configureOutput(nodeId, output.SinkId, output.Id, output.Config)

	return nil
}

func (r *Registry) configureOutput(nodeId, sinkId, outputId uuid.UUID, cfg json.RawMessage) {
	err := r.writeNode(nodeId, event.ConfigureOutput{
		SinkId:   sinkId,
		OutputId: outputId,
		Config:   cfg,
	})
	if err != nil {
		fmt.Println("error sending event:", err)
		return
	}
}

func (r *Registry) handleOutputDisconnected(c Conn, e event.OutputDisconnected) error {
	if _, ok := r.Conns[c.Addr()]; !ok {
		return errors.New("not connected")
	}

	nodeId, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("not authenticated")
	}

	node := r.State.Nodes[nodeId]

	sink, ok := r.State.Sinks[e.SinkId]
	if !ok {
		return errors.New("unknown sink")
	}

	if sink.NodeId != node.Id {
		return errors.New("sink not found")
	}

	output, ok := r.State.Outputs[e.Id]
	if !ok {
		return errors.New("unknown output")
	}

	if output.SinkId != e.SinkId {
		return errors.New("output not found")
	}

	output.Connected = false

	return nil
}

func (r *Registry) handleData(c Conn, e event.Data) error {
	_, ok := r.AddrNode[c.Addr()]
	if !ok {
		return errors.New("node disconnected")
	}

	//node := r.State.Nodes[nodeId]

	dest := r.State.Nodes[e.NodeId]
	if dest == nil {
		return fmt.Errorf("unknown node: %s", e.NodeId)
	}

	// TODO: verify this node is actually sending to the dest

	srcOutputs := r.activeNodeOutputs(dest.Id)

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

	go func() {
		addr, ok := r.NodeAddr[dest.Id]
		if !ok {
			return
		}

		c, ok := r.Conns[addr]
		if !ok {
			return
		}

		err := r.write(c, e)
		if err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}
