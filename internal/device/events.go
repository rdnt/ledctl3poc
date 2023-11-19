package device

import (
	"fmt"

	"ledctl3/event"
)

func (s *Device) ProcessEvent(addr string, e event.Event) {
	s.mux.Lock()
	defer s.mux.Unlock()

	//fmt.Println("ProcessEvents")

	switch e := e.(type) {
	case event.Connect:
		s.handleConnectEvent(addr, e)
	case event.Disconnect:
		s.handleDisconnectEvent(addr, e)
	//case event.ListCapabilities:
	//	s.handleListCapabilitiesEvent(addr, e)
	default:
		fmt.Println("unknown event", e)
	}

	//fmt.Println("ProcessEvents done")
}

func (s *Device) handleConnectEvent(addr string, e event.Connect) {
	fmt.Printf("%s: recv ListCapabilities\n", addr)

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
			Id:     in.Id(),
			Type:   event.InputTypeDefault,
			Schema: in.Schema(),
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

func (s *Device) handleDisconnectEvent(addr string, _ event.Disconnect) {
	fmt.Printf("%s: recv Disconnect\n", addr)

	s.regAddr = ""
}

//func (s *Device) handleListCapabilitiesEvent(addr string, _ event.ListCapabilities) {
//	fmt.Printf("%s: recv ListCapabilities\n", addr)
//
//	e := event.Capabilities{
//		Inputs: lo.Map(lo.Values(s.inputs), func(input common.Input, _ int) event.CapabilitiesInput {
//			return event.CapabilitiesInput{
//				Id:     input.Id(),
//				Type:   event.InputTypeDefault,
//				Schema: input.Schema(),
//			}
//		}),
//		Outputs: lo.Map(lo.Values(s.outputs), func(out common.Output, _ int) event.CapabilitiesOutput {
//			return event.CapabilitiesOutput{
//				Id:   out.Id(),
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
