package node

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/node/event"
	"ledctl3/node/types"
)

type ConnectedEvent struct{}

func (e ConnectedEvent) Type() string {
	return "Connected"
}

type DisconnectedEvent struct{}

func (e DisconnectedEvent) Type() string {
	return "Disconnected"
}

type Conn interface {
	Addr() string
	Write(b []byte) error
	Read() ([]byte, error)
	Ping() error
}

func (n *Node) ProcessEvent(c Conn, e any) error {
	n.mux.Lock()
	defer n.mux.Unlock()

	//fmt.Println("HandleConnection")

	if _, ok := e.(event.Data); !ok {
		fmt.Printf("%s: recv %s\n", c.Addr(), reflect.TypeOf(e).Name())
	}

	var err error

	switch e := e.(type) {
	case ConnectedEvent:
		n.handleConnected(c)
	case DisconnectedEvent:
		n.handleDisconnected(c)
	//case event.ConfigureSource:
	//	n.handleConfigureSource(addr, e)
	case event.ConfigureSource:
		n.handleConfigureSource(c, e)

	case event.ConfigureSink:
		err = n.handleConfigureSink(c, e)

	//case event.SetSourceConfig:
	//	err = n.handleSetSourceConfig(c, e)
	//case event.SetSinkConfig:
	//	err = n.handleSetSinkConfig(c, e)
	//
	case event.ConfigureInput:
		err = n.handleConfigureInput(c, e)
	case event.ConfigureOutput:
		err = n.handleConfigureOutput(c, e)

	case event.ConfigureLinks:
		err = n.handleConfigureLinks(c, e)

	case event.Data:
		n.handleData(c, e)
	//case event.ListCapabilities:
	//	n.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	if _, ok := e.(event.Data); !ok {
		fmt.Printf("%s: proc %s\n", c.Addr(), reflect.TypeOf(e).Name())
	}

	return err
}

func (n *Node) handleConfigureInput(c Conn, e event.ConfigureInput) error {
	src, ok := n.sources[e.SourceId]
	if !ok {
		return errors.New("source not found")
	}

	in, ok := n.inputs[e.InputId]
	if !ok {
		return errors.New("input not found")
	}

	if in.SourceId() != src.Id() {
		return errors.New("invalid input source")
	}

	err := in.Configure(e.Config)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) handleConfigureLinks(c Conn, e event.ConfigureLinks) error {
	src, ok := n.sources[e.SourceId]
	if !ok {
		return errors.New("source not found")
	}

	in, ok := n.inputs[e.InputId]
	if !ok {
		return errors.New("input not found")
	}

	if in.SourceId() != src.Id() {
		return errors.New("invalid input source")
	}

	var cfgs []types.OutputConfig

	for _, cfg := range e.Outputs {
		cfgs = append(cfgs, types.OutputConfig{
			OutputId: cfg.OutputId,
			NodeId:   cfg.NodeId,
			SinkId:   cfg.SinkId,
			Leds:     cfg.Leds,
			Config:   cfg.Config,
		})
	}

	err := in.Connect(cfgs)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) handleConfigureOutput(c Conn, e event.ConfigureOutput) error {
	sink, ok := n.sinks[e.SinkId]
	if !ok {
		return errors.New("source not found")
	}

	out, ok := n.outputs[e.OutputId]
	if !ok {
		return errors.New("input not found")
	}

	if out.SinkId() != sink.Id() {
		return errors.New("invalid input source")
	}

	err := out.Configure(e.Config)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) write(c Conn, v event.Event) error {
	codec := event.JSONCodec{}

	if _, ok := v.(event.Data); !ok {
		fmt.Printf("%s: send %s\n", c.Addr(), reflect.TypeOf(v).Name())
	}

	b, err := codec.MarshalBinary(v)
	if err != nil {
		return err
	}

	if c == nil {
		panic("conn is nil")
	}

	return c.Write(b)
}

func (n *Node) handleConnected(c Conn) {
	err := n.write(c, event.Authenticate{
		Id: n.cfg.Id,
	})
	if err != nil {
		fmt.Println("error writing to addr", c.Addr(), err)
	}

	var sources []event.CapabilitiesSource
	for _, d := range n.sources {
		sources = append(sources, event.CapabilitiesSource{
			Id:     d.Id(),
			Schema: d.Schema(),
		})
	}

	var sinks []event.CapabilitiesSink
	for _, d := range n.sinks {
		sinks = append(sinks, event.CapabilitiesSink{
			Id:     d.Id(),
			Schema: d.Schema(),
		})
	}

	err = n.write(c, event.Capabilities{
		//OutputId:      n.cfg.OutputId,
		Sources: sources,
		Sinks:   sinks,
	})
	if err != nil {
		fmt.Println("error writing to addr", c.Addr(), err)
	}

	for _, in := range n.inputs {
		err := n.write(c, event.InputConnected{
			Id:       in.Id(),
			SourceId: in.SourceId(),
			//Type:   event.InputTypeDefault,
			//Schema: in.Schema(),
		})
		if err != nil {
			fmt.Println("error writing to addr", c.Addr(), err)
			return
		}
	}

	for _, out := range n.outputs {
		err := n.write(c, event.OutputConnected{
			Id:     out.Id(),
			SinkId: out.SinkId(),
			Leds:   out.Leds(),
		})
		if err != nil {
			fmt.Println("error writing to addr", c.Addr(), err)
			return
		}
	}

	n.connMux.Lock()
	n.conn = c
	n.connMux.Unlock()
}

func (n *Node) handleDisconnected(c Conn) {
	n.connMux.Lock()
	n.conn = nil
	n.connMux.Unlock()
}

//func (s *Node) handleListCapabilitiesEvent(addr string, _ event.ListCapabilities) {
//	fmt.Printf("%s: recv ListCapabilities\n", addr)
//
//	e := event.Capabilities{
//		Inputs: lo.Map(lo.Values(s.inputs), func(input common.Input, _ int) event.CapabilitiesInput {
//			return event.CapabilitiesInput{
//				OutputId:     input.OutputId(),
//				Type:   event.InputTypeDefault,
//				Schema: input.Schema(),
//			}
//		}),
//		Outputs: lo.Map(lo.Values(s.outputs), func(out common.Output, _ int) event.CapabilitiesOutput {
//			return event.CapabilitiesOutput{
//				OutputId:   out.OutputId(),
//				Leds: out.Leds(),
//			}
//		}),
//	}
//
//	fmt.Printf("%s: send Capabilities\n", addr)
//	err := s.write(addr, e)
//	if err != nil {
//		fmt.Println("error writing to addr", addr, err)
//	}
//}

//func (n *Node) handleConfigureSource(addr string, e event.ConfigureSource) {
//	fmt.Printf("%s: recv ConfigureSource\n", addr)
//
//	b, err := json.Marshal(e)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(string(b))
//
//	for _, input := range e.Inputs {
//		in, ok := n.inputs[input.OutputId]
//		if !ok {
//			fmt.Println("in not found", input.OutputId)
//			continue
//		}
//
//		var outputCfgs []types.OutputConfig
//		for _, output := range input.Outputs {
//			outputCfgs = append(outputCfgs, types.OutputConfig{
//				OutputId: output.OutputId,
//				//SinkId: output.SinkId,
//				SinkId: output.SinkId,
//				Config: types.LinkConfig{},
//				Leds:   output.Leds,
//			})
//		}
//
//		err := in.Init(types.LinksConfig{
//			Framerate: 60,
//			Outputs:   outputCfgs,
//		})
//		if err != nil {
//			fmt.Println(err)
//			continue
//		}
//
//		fmt.Println("input started", input.OutputId)
//	}
//}

func (n *Node) handleConfigureSource(c Conn, e event.ConfigureSource) {
	src, ok := n.sources[e.SourceId]
	if !ok {
		fmt.Println("source not found", e.SourceId)
		return
	}

	//inputCfgs := make(map[uuid.UUID]types.InputConfig)
	//
	//for _, e := range e.Inputs {
	//	in, ok := n.inputs[e.InputId]
	//	if !ok {
	//		fmt.Println("input not found", e.InputId)
	//		continue
	//	}
	//
	//	outputCfgs := make([]types.OutputConfig, 0, len(e.Outputs))
	//
	//	for _, output := range e.Outputs {
	//		outputCfgs = append(outputCfgs, types.OutputConfig{
	//			NodeId:   output.NodeId,
	//			SinkId:   output.SinkId,
	//			OutputId: output.OutputId,
	//			Leds:     output.Leds,
	//			Config:   output.Config,
	//		})
	//	}
	//
	//	inputCfgs[in.Id()] = types.InputConfig{
	//		InputId: e.InputId,
	//		Config:  e.InputConfig,
	//		Outputs: outputCfgs,
	//	}
	//}

	err := src.Configure(e.SourceConfig)
	if err != nil {
		fmt.Println(err)
	}
}

func (n *Node) handleData(c Conn, e event.Data) {
	syncOuts := map[uuid.UUID][]event.DataOutput{}

	for _, out := range e.Outputs {
		sink, ok := n.sinks[out.SinkId]
		if !ok {
			continue
		}

		if _, ok := n.outputs[out.OutputId]; !ok {
			continue
		}

		syncOuts[sink.Id()] = append(syncOuts[sink.Id()], out)
	}

	for sinkId, outs := range syncOuts {
		sink := n.sinks[sinkId]

		pixs := map[uuid.UUID][]color.NRGBA{}

		for _, out := range outs {
			pixs[out.OutputId] = out.Pix
		}

		err := sink.Render(pixs)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

//func (n *Node) handleSetSourceConfig(c Conn, e event.SetSourceConfig) error {
//	//@@@@@@@@@@@@@@@@@@@@@
//	//dev, ok := n.sources[e.SourceId]
//	//if !ok {
//	//	return errors.New("source not found")
//	//}
//
//	//err := dev.Configure(e.Config)
//	//if err != nil {
//	//	return err
//	//}
//
//	fmt.Println("source config set", e.SourceId)
//	return nil
//}

//func (n *Node) handleSetSinkConfig(c Conn, e event.SetSinkConfig) error {
//	dev, ok := n.sinks[e.SinkId]
//	if !ok {
//		return errors.New("sink not found")
//	}
//
//	err := dev.Configure(e.Config)
//	if err != nil {
//		return err
//	}
//
//	fmt.Println("sink config set", e.SinkId)
//	return nil
//}
