package device

import (
	"fmt"
	"sync"

	"ledctl3/event"
	"ledctl3/internal/device/common"
	"ledctl3/pkg/uuid"
)

type Client struct {
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

func New(cfg Config, write func(addr string, e event.Event) error) (*Client, error) {
	d := &Client{
		id:      cfg.Id,
		write:   write,
		cfg:     cfg,
		inputs:  make(map[uuid.UUID]common.Input),
		outputs: make(map[uuid.UUID]common.Output),
	}

	for _, driver := range devices {
		err := driver.Start(d) // TODO: stop devices if one fails to start
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (s *Client) AddInput(in common.Input) {
	//fmt.Println("ADD INPUT CALLED", in)

	s.inputs[in.Id()] = in
	//s.inputCfgs[in.OutputId()] = inputConfig{}

	go func() {
		// forward messages from input to the network
		for e := range in.Events() {

			// TODO: synchronized render won't work for same-device i/o.
			//  possible solution: instead of sending data to registry, send
			//  directly to sink device, and calculate and send to registry
			//  the RTT from source to sink. the registry can then calculate
			//  how much the sink should offset its render time to match the
			//  latency of the slowest device on the network.
			//if e.SinkId == s.id {
			//	// deliver to local device outputs
			//
			//	for _, out := range e.Outputs {
			//		if _, ok := s.outputs[out.OutputId]; !ok {
			//			fmt.Println("output not found", out.OutputId)
			//			continue
			//		}
			//
			//		s.outputs[out.OutputId].Render(out.Pix)
			//	}
			//
			//	continue
			//}

			e := e
			go func() {
				var outputs []event.DataOutput
				for _, output := range e.Outputs {
					outputs = append(outputs, event.DataOutput{
						OutputId: output.OutputId,
						Pix:      output.Pix,
					})
				}

				if s.regAddr == "" {
					return
				}

				err := s.write(s.regAddr, event.Data{
					SinkId:  e.SinkId,
					Outputs: outputs,
					Latency: e.Latency,
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
			}()

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

type Device interface {
	Start(reg common.IORegistry) error
	SetConfig(cfg []byte) error
	Schema() ([]byte, error)
	Config() ([]byte, error)
}

var devices []Device
var devicesMux sync.Mutex

func Register(driver Device) {
	devicesMux.Lock()
	defer devicesMux.Unlock()

	devices = append(devices, driver)
}

func (s *Client) RemoveInput(id uuid.UUID) {
	//fmt.Println("RemoveInput CALLED", id)
	delete(s.inputs, id)
}

func (s *Client) AddOutput(out common.Output) {
	//fmt.Println("ADD OUTPUT CALLED", out)

	s.outputs[out.Id()] = out
}

func (s *Client) RemoveOutput(id uuid.UUID) {
	//fmt.Println("REMOVE OUTPUT CALLED", id)

	delete(s.outputs, id)
}

func (s *Client) handleData(addr string, e event.Data) {
	for _, out := range e.Outputs {
		if _, ok := s.outputs[out.OutputId]; !ok {
			fmt.Println("output not found", out.OutputId)
			continue
		}

		go s.outputs[out.OutputId].Render(out.Pix)
	}
}
