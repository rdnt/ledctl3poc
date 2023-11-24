package device

import (
	"fmt"
	"sync"

	"ledctl3/event"
	"ledctl3/internal/device/common"
	"ledctl3/pkg/uuid"
)

type Device struct {
	id      uuid.UUID
	mux     sync.Mutex
	write   func(addr string, e event.Event) error
	cfg     Config
	inputs  map[uuid.UUID]common.Input
	outputs map[uuid.UUID]common.Output
	regAddr string
}

type Config struct {
	Id uuid.UUID
}

func New(cfg Config, write func(addr string, e event.Event) error) (*Device, error) {
	return &Device{
		id:      cfg.Id,
		write:   write,
		cfg:     cfg,
		inputs:  make(map[uuid.UUID]common.Input),
		outputs: make(map[uuid.UUID]common.Output),
	}, nil
}

func (s *Device) AddInput(in common.Input) {
	//fmt.Println("ADD INPUT CALLED", in)

	s.inputs[in.Id()] = in
	//s.inputCfgs[in.Id()] = inputConfig{}

	go func() {
		// forward messages from input to the network
		for e := range in.Events() {

			if e.SinkId == s.id {
				// deliver to local device outputs

				for _, out := range e.Outputs {
					if _, ok := s.outputs[out.OutputId]; !ok {
						fmt.Println("output not found", out.OutputId)
						continue
					}

					s.outputs[out.OutputId].Render(out.Pix)
				}

				continue
			}

			var outputs []event.DataOutput
			for _, output := range e.Outputs {
				outputs = append(outputs, event.DataOutput{
					Id:  output.OutputId,
					Pix: output.Pix,
				})
			}

			if s.regAddr == "" {
				continue
			}

			err := s.write(s.regAddr, event.Data{
				SinkId:  e.SinkId,
				Outputs: outputs,
			})
			if err != nil {
				fmt.Println("write error:", err)
			}

			//s.messages <- Message{
			//	Addr: nil, // TODO: registry addr
			//	Event: event.Data{
			//		Event:     event.Event{Type: event.Data},
			//		SessionId: s.sessionId,
			//		Outputs:   outputs,
			//	},
			//}
		}
	}()
}

func (s *Device) RemoveInput(id uuid.UUID) {
	//fmt.Println("RemoveInput CALLED", id)
	delete(s.inputs, id)
}

func (s *Device) AddOutput(out common.Output) {
	//fmt.Println("ADD OUTPUT CALLED", out)

	s.outputs[out.Id()] = out
}

func (s *Device) RemoveOutput(id uuid.UUID) {
	//fmt.Println("REMOVE OUTPUT CALLED", id)

	delete(s.outputs, id)
}

func (s *Device) handleData(addr string, e event.Data) {
	for _, out := range e.Outputs {
		if _, ok := s.outputs[out.Id]; !ok {
			fmt.Println("output not found", out.Id)
			continue
		}

		s.outputs[out.Id].Render(out.Pix)
	}
}
