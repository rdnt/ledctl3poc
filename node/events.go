package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
	Latency() time.Duration
	Write(b []byte) error
}

func (n *Node) ProcessEvent(c Conn, e any) error {
	n.mux.Lock()
	defer n.mux.Unlock()

	//fmt.Println("HandleConnection")

	var err error

	switch e := e.(type) {
	case ConnectedEvent:
		n.handleConnected(c)
	case DisconnectedEvent:
		n.handleDisconnected(c)
	//case event.SetSourceActive:
	//	n.handleSetSourceActive(addr, e)
	case event.SetInputActive:
		n.handleSetInputActive(c, e)
	case event.SetSourceConfig:
		err = n.handleSetSourceConfig(c, e)
	case event.SetSinkConfig:
		err = n.handleSetSinkConfig(c, e)
	case event.Data:
		n.handleData(c, e)
	//case event.ListCapabilities:
	//	n.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	//fmt.Println("HandleConnection done")

	return err
}

func (n *Node) write(c Conn, v event.Event) error {
	codec := event.JSONCodec{}

	b, err := codec.MarshalBinary(v)
	if err != nil {
		return err
	}

	return c.Write(b)
}

func (n *Node) handleConnected(c Conn) {
	fmt.Printf("%s: recv Connected\n", c.Addr())

	fmt.Printf("%s: send Authenticate\n", c.Addr())
	err := n.write(c, event.Authenticate{
		Id: n.cfg.Id,
	})
	if err != nil {
		fmt.Println("error writing to addr", c.Addr(), err)
	}

	var sources []event.ConnectedSource
	for _, d := range n.sources {
		cfg, err := d.Config()
		if err != nil {
			fmt.Println(err)
			return
		}

		schema, err := d.Schema()
		if err != nil {
			fmt.Println(err)
			return
		}

		sources = append(sources, event.ConnectedSource{
			Id:     d.Id(),
			Config: cfg,
			Schema: schema,
		})
	}

	var sinks []event.ConnectedSink
	for _, d := range n.sinks {
		cfg, err := d.Config()
		if err != nil {
			fmt.Println(err)
			return
		}

		schema, err := d.Schema()
		if err != nil {
			fmt.Println(err)
			return
		}

		sinks = append(sinks, event.ConnectedSink{
			Id:     d.Id(),
			Config: cfg,
			Schema: schema,
		})
	}

	fmt.Printf("%s: send Capabilities\n", c.Addr())

	err = n.write(c, event.Capabilities{
		//Id:      n.cfg.Id,
		Sources: sources,
		Sinks:   sinks,
	})
	if err != nil {
		fmt.Println("error writing to addr", c.Addr(), err)
	}

	for _, in := range n.inputs {
		fmt.Printf("%s: send InputConnected\n", c.Addr())

		err := n.write(c, event.InputConnected{
			Id:       in.Id(),
			DriverId: in.DriverId(),
			//Type:   event.InputTypeDefault,
			//Schema: in.Schema(),
		})
		if err != nil {
			fmt.Println("error writing to addr", c.Addr(), err)
			return
		}
	}

	for _, out := range n.outputs {
		fmt.Printf("%s: send OutputConnected\n", c.Addr())

		err := n.write(c, event.OutputConnected{
			Id:       out.Id(),
			DriverId: out.DriverId(),
			Leds:     out.Leds(),
		})
		if err != nil {
			fmt.Println("error writing to addr", c.Addr(), err)
			return
		}
	}

	n.regAddr = c.Addr()
}

func (n *Node) handleDisconnected(c Conn) {
	fmt.Printf("%s: recv Disconnect\n", c.Addr())

	n.regAddr = ""
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

func (n *Node) handleSetSourceActive(addr string, e event.SetSourceActive) {
	fmt.Printf("%s: recv SetSourceActive\n", addr)

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	for _, input := range e.Inputs {
		in, ok := n.inputs[input.Id]
		if !ok {
			fmt.Println("in not found", input.Id)
			continue
		}

		var outputCfgs []types.OutputConfig
		for _, output := range input.Outputs {
			outputCfgs = append(outputCfgs, types.OutputConfig{
				Id: output.Id,
				//DriverId: output.DriverId,
				SinkId: output.SinkId,
				Config: types.OutputConfigConfig{},
				Leds:   output.Leds,
			})
		}

		err := in.Start(types.InputConfig{
			Framerate: 60,
			Outputs:   outputCfgs,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("input started", input.Id)
	}
}

func (n *Node) handleSetInputActive(c Conn, e event.SetInputActive) {
	fmt.Printf("%s: recv SetInputActive\n", c.Addr())

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	in, ok := n.inputs[e.Id]
	if !ok {
		fmt.Println("input not found", e.Id)
		return
	}

	var outputCfgs []types.OutputConfig
	for _, output := range e.Outputs {
		var outCfg types.OutputConfigConfig
		b, err := json.Marshal(output.Config)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, &outCfg)
		if err != nil {
			panic(err)
		}

		outputCfgs = append(outputCfgs, types.OutputConfig{
			Id:     output.OutputId,
			SinkId: output.SinkId,
			NodeId: n.id,
			Leds:   output.Leds,
			Config: outCfg,
		})
	}

	err = in.Start(types.InputConfig{
		Framerate: 60,
		Outputs:   outputCfgs,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("input started", e.Id)
}

func (n *Node) handleData(c Conn, e event.Data) {
	for _, out := range e.Outputs {
		if _, ok := n.outputs[out.OutputId]; !ok {
			fmt.Println("output not found", out.OutputId)
			continue
		}

		go n.outputs[out.OutputId].Render(out.Pix)
	}
}

func (n *Node) handleSetSourceConfig(c Conn, e event.SetSourceConfig) error {
	fmt.Printf("%s: recv SetSourceConfig\n", c.Addr())

	dev, ok := n.sources[e.SourceId]
	if !ok {
		return errors.New("source not found")
	}

	err := dev.SetConfig(e.Config)
	if err != nil {
		return err
	}

	fmt.Println("source config set", e.SourceId)
	return nil
}

func (n *Node) handleSetSinkConfig(c Conn, e event.SetSinkConfig) error {
	fmt.Printf("%s: recv SetSinkConfig\n", c.Addr())

	dev, ok := n.sinks[e.SinkId]
	if !ok {
		return errors.New("sink not found")
	}

	err := dev.SetConfig(e.Config)
	if err != nil {
		return err
	}

	fmt.Println("sink config set", e.SinkId)
	return nil
}
