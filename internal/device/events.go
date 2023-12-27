package device

import (
	"encoding/json"
	"fmt"

	"ledctl3/event"
	"ledctl3/internal/device/types"
)

func (s *Client) ProcessEvent(addr string, e event.Event) {
	s.mux.Lock()
	defer s.mux.Unlock()

	//fmt.Println("ProcessEvents")

	switch e := e.(type) {
	case event.Connect:
		s.handleConnect(addr, e)
	case event.Disconnect:
		s.handleDisconnect(addr, e)
	//case event.SetSourceActive:
	//	s.handleSetSourceActive(addr, e)
	case event.SetInputActive:
		s.handleSetInputActive(addr, e)
	case event.Data:
		s.handleData(addr, e)
	//case event.ListCapabilities:
	//	s.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Printf("unknown event %#v\n", e)
	}

	//fmt.Println("ProcessEvents done")
}

func (s *Client) handleConnect(addr string, e event.Connect) {
	fmt.Printf("%s: recv Connect\n", addr)

	fmt.Printf("%s: send Connect\n", addr)
	err := s.write(addr, event.Connect{
		Id: s.cfg.Id,
	})
	if err != nil {
		fmt.Println("error writing to addr", addr, err)
	}

	for _, in := range s.inputs {
		fmt.Printf("%s: send InputConnected\n", addr)

		err := s.write(addr, event.InputConnected{
			Id: in.Id(),
			//Type:   event.InputTypeDefault,
			//Schema: in.Schema(),
		})
		if err != nil {
			fmt.Println("error writing to addr", addr, err)
			return
		}
	}

	for _, out := range s.outputs {
		fmt.Printf("%s: send OutputConnected\n", addr)

		err := s.write(addr, event.OutputConnected{
			Id:   out.Id(),
			Leds: out.Leds(),
		})
		if err != nil {
			fmt.Println("error writing to addr", addr, err)
			return
		}
	}

	s.regAddr = addr
}

func (s *Client) handleDisconnect(addr string, _ event.Disconnect) {
	fmt.Printf("%s: recv Disconnect\n", addr)

	s.regAddr = ""
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

func (s *Client) handleSetSourceActive(addr string, e event.SetSourceActive) {
	fmt.Printf("%s: recv SetSourceActive\n", addr)

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	for _, input := range e.Inputs {
		in, ok := s.inputs[input.Id]
		if !ok {
			fmt.Println("in not found", input.Id)
			continue
		}

		var outputCfgs []types.OutputConfig
		for _, output := range input.Outputs {
			outputCfgs = append(outputCfgs, types.OutputConfig{
				Id:     output.Id,
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

func (s *Client) handleSetInputActive(addr string, e event.SetInputActive) {
	fmt.Printf("%s: recv SetInputActive\n", addr)

	b, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	in, ok := s.inputs[e.Id]
	if !ok {
		fmt.Println("in not found", e.Id)
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
