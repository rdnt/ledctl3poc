package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	drivers map[uuid.UUID]Driver
	regAddr string
}

type Config struct {
	Id uuid.UUID
}

type stateHolder struct {
	mux        sync.Mutex
	driverId   uuid.UUID
	driverName string
}

type driverConfig struct {
	Id     uuid.UUID       `json:"id"`
	Config json.RawMessage `json:"config"`
	State  json.RawMessage `json:"state"`
}

func (s *stateHolder) Id() uuid.UUID {
	return s.driverId
}

func (s *stateHolder) getId() (uuid.UUID, error) {
	cfg, err := s.get()
	if errors.Is(err, os.ErrNotExist) {
		cfg = driverConfig{
			Id:     uuid.New(),
			Config: json.RawMessage("null"),
			State:  json.RawMessage("null"),
		}

		err = s.set(cfg)
		if err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	}

	s.driverId = cfg.Id
	return cfg.Id, nil
}

func (s *stateHolder) SetConfig(b []byte) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}

	cfg.Config = b

	return s.set(cfg)
}

func (s *stateHolder) get() (driverConfig, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	b, err := os.ReadFile(fmt.Sprintf("./%s-driver.json", s.driverName))
	if err != nil {
		return driverConfig{}, err
	}

	var cfg driverConfig
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return driverConfig{}, err
	}

	return cfg, nil

}

func (s *stateHolder) GetConfig() ([]byte, error) {
	cfg, err := s.get()
	if err != nil {
		return nil, err
	}

	return cfg.Config, nil
}

func (s *stateHolder) SetState(b []byte) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}

	cfg.State = b

	return s.set(cfg)
}

func (s *stateHolder) GetState() ([]byte, error) {
	cfg, err := s.get()
	if err != nil {
		return nil, err
	}

	return cfg.State, nil
}

func (s *stateHolder) set(cfg driverConfig) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("./%s-driver.json", s.driverName), b, 0644)
}

func newDriverStateHolder(name string) *stateHolder {
	return &stateHolder{driverName: name}
}

func New(cfg Config, write func(addr string, e event.Event) error) (*Client, error) {
	d := &Client{
		id:      cfg.Id,
		write:   write,
		cfg:     cfg,
		inputs:  make(map[uuid.UUID]common.Input),
		outputs: make(map[uuid.UUID]common.Output),
		drivers: make(map[uuid.UUID]Driver),
	}

	for name, drv := range drivers {
		sh := newDriverStateHolder(name)
		id, err := sh.getId()
		if err != nil {
			return nil, err
		}

		err = drv.Start(id, d, sh) // TODO: stop drivers if one fails to start
		if err != nil {
			return nil, err
		}

		d.drivers[id] = drv
	}

	return d, nil
}

func (c *Client) AddInput(in common.Input) {
	//fmt.Println("ADD INPUT CALLED", in)

	c.inputs[in.Id()] = in
	//c.inputCfgs[in.OutputId()] = inputConfig{}

	go func() {
		// forward messages from input to the network
		for e := range in.Events() {

			// TODO: synchronized render won't work for same-device i/o.
			//  possible solution: instead of sending data to registry, send
			//  directly to sink device, and calculate and send to registry
			//  the RTT from source to sink. the registry can then calculate
			//  how much the sink should offset its render time to match the
			//  latency of the slowest device on the network.
			//if e.DriverId == c.id {
			//	// deliver to local device outputs
			//
			//	for _, out := range e.Outputs {
			//		if _, ok := c.outputs[out.OutputId]; !ok {
			//			fmt.Println("output not found", out.OutputId)
			//			continue
			//		}
			//
			//		c.outputs[out.OutputId].Render(out.Pix)
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

				if c.regAddr == "" {
					return
				}

				err := c.write(c.regAddr, event.Data{
					SinkId:  e.SinkId,
					Outputs: outputs,
					Latency: e.Latency,
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
			}()

			//c.messages <- Message{
			//	Addr: nil, // TODO: registry addr
			//	Event: event.Data{
			//		Event:     event.Event{Type: event.Data},
			//		SessionId: c.sessionId,
			//		Outputs:   outputs,
			//	},
			//}
		}
	}()
}

type Driver interface {
	Id() uuid.UUID
	Start(id uuid.UUID, reg common.IORegistry, stateHolder common.StateHolder) error
	SetConfig(cfg []byte) error
	Schema() ([]byte, error)
	Config() ([]byte, error)
}

var drivers = map[string]Driver{}
var driversMux sync.Mutex

func Register(name string, driver Driver) {
	driversMux.Lock()
	defer driversMux.Unlock()

	drivers[name] = driver
}

func (c *Client) RemoveInput(id uuid.UUID) {
	//fmt.Println("RemoveInput CALLED", id)
	delete(c.inputs, id)
}

func (c *Client) AddOutput(out common.Output) {
	//fmt.Println("ADD OUTPUT CALLED", out)

	c.outputs[out.Id()] = out
}

func (c *Client) RemoveOutput(id uuid.UUID) {
	//fmt.Println("REMOVE OUTPUT CALLED", id)

	delete(c.outputs, id)
}
