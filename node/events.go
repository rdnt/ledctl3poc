package node

import (
	"encoding/json"
	"fmt"

	"ledctl3/node/event"
	"ledctl3/node/types"
)

type ConnectedEvent struct{}

type DisconnectedEvent struct{}

func (c *Client) ProcessEvent(addr string, e any) {
	c.mux.Lock()
	defer c.mux.Unlock()

	//fmt.Println("HandleConnection")

	switch e := e.(type) {
	case ConnectedEvent:
		c.handleConnected(addr)
	case DisconnectedEvent:
		c.handleDisconnected(addr)
	//case event.SetSourceActive:
	//	c.handleSetSourceActive(addr, e)
	case event.SetInputActive:
		c.handleSetInputActive(addr, e)
	case event.SetSourceConfig:
		c.handleSetSourceConfig(addr, e)
	case event.SetSinkConfig:
		c.handleSetSinkConfig(addr, e)
	case event.Data:
		c.handleData(addr, e)
	//case event.ListCapabilities:
	//	c.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	//fmt.Println("HandleConnection done")
}

func (c *Client) handleConnected(addr string) {
	fmt.Printf("%s: recv Connected\n", addr)

	var sources []event.ConnectedSource
	for _, d := range c.sources {
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
	for _, d := range c.sinks {
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

	fmt.Printf("%s: send NodeConnected\n", addr)
	err := c.write(addr, event.NodeConnected{
		Id:      c.cfg.Id,
		Sources: sources,
		Sinks:   sinks,
	})
	if err != nil {
		fmt.Println("error writing to addr", addr, err)
	}

	for _, in := range c.inputs {
		fmt.Printf("%s: send InputConnected\n", addr)

		err := c.write(addr, event.InputConnected{
			Id:       in.Id(),
			DriverId: in.DriverId(),
			//Type:   event.InputTypeDefault,
			//Schema: in.Schema(),
		})
		if err != nil {
			fmt.Println("error writing to addr", addr, err)
			return
		}
	}

	for _, out := range c.outputs {
		fmt.Printf("%s: send OutputConnected\n", addr)

		err := c.write(addr, event.OutputConnected{
			Id:       out.Id(),
			DriverId: out.DriverId(),
			Leds:     out.Leds(),
		})
		if err != nil {
			fmt.Println("error writing to addr", addr, err)
			return
		}
	}

	c.regAddr = addr
}

func (c *Client) handleDisconnected(addr string) {
	fmt.Printf("%s: recv Disconnect\n", addr)

	c.regAddr = ""
}

//func (s *Client) handleListCapabilitiesEvent(addr string, _ event.ListCapabilities) {
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

func (c *Client) handleSetSourceActive(addr string, e event.SetSourceActive) {
	fmt.Printf("%s: recv SetSourceActive\n", addr)

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	for _, input := range e.Inputs {
		in, ok := c.inputs[input.Id]
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

func (c *Client) handleSetInputActive(addr string, e event.SetInputActive) {
	fmt.Printf("%s: recv SetInputActive\n", addr)

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	in, ok := c.inputs[e.Id]
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
			NodeId: c.id,
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

func (c *Client) handleData(addr string, e event.Data) {
	for _, out := range e.Outputs {
		if _, ok := c.outputs[out.OutputId]; !ok {
			fmt.Println("output not found", out.OutputId)
			continue
		}

		go c.outputs[out.OutputId].Render(out.Pix)
	}
}

func (c *Client) handleSetSourceConfig(addr string, e event.SetSourceConfig) {
	fmt.Printf("%s: recv SetSourceConfig\n", addr)

	dev, ok := c.sources[e.SourceId]
	if !ok {
		fmt.Println("source not found", e.SourceId)
		return
	}

	err := dev.SetConfig(e.Config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("source config set", e.SourceId)
}

func (c *Client) handleSetSinkConfig(addr string, e event.SetSinkConfig) {
	fmt.Printf("%s: recv SetSinkConfig\n", addr)

	dev, ok := c.sinks[e.SinkId]
	if !ok {
		fmt.Println("sink not found", e.SinkId)
		return
	}

	err := dev.SetConfig(e.Config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("sink config set", e.SinkId)
}
