package node

import (
	"encoding/json"
	"fmt"

	"ledctl3/event"
	"ledctl3/node/types"
)

func (c *Client) ProcessEvent(addr string, e event.Event) {
	c.mux.Lock()
	defer c.mux.Unlock()

	//fmt.Println("ProcessEvents")

	switch e := e.(type) {
	case event.Connect:
		c.handleConnect(addr, e)
	case event.Disconnect:
		c.handleDisconnect(addr, e)
	//case event.SetSourceActive:
	//	c.handleSetSourceActive(addr, e)
	case event.SetInputActive:
		c.handleSetInputActive(addr, e)
	case event.SetDriverConfig:
		c.handleSetDeviceConfig(addr, e)
	case event.Data:
		c.handleData(addr, e)
	//case event.ListCapabilities:
	//	c.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	//fmt.Println("ProcessEvents done")
}

func (c *Client) handleConnect(addr string, e event.Connect) {
	fmt.Printf("%s: recv Connect\n", addr)

	var drivers []event.ConnectDriver
	for _, d := range c.drivers {
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

		drivers = append(drivers, event.ConnectDriver{
			Id:     d.Id(),
			Config: cfg,
			Schema: schema,
		})
	}

	fmt.Printf("%s: send Connect\n", addr)
	err := c.write(addr, event.Connect{
		Id:      c.cfg.Id,
		Drivers: drivers,
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

func (c *Client) handleDisconnect(addr string, _ event.Disconnect) {
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
			Id:       output.OutputId,
			SinkId:   output.SinkId,
			DeviceId: c.id,
			Leds:     output.Leds,
			Config:   outCfg,
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

func (c *Client) handleSetDeviceConfig(addr string, e event.SetDriverConfig) {
	fmt.Printf("%s: recv SetDriverConfig\n", addr)

	dev, ok := c.drivers[e.DriverId]
	if !ok {
		fmt.Println("device not found", e.DriverId)
		return
	}

	err := dev.SetConfig(e.Config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("device config set", e.DriverId)
}
