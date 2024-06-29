package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"os"
	"sync"

	"github.com/google/uuid"

	"ledctl3/node/common"
	"ledctl3/node/event"
)

type Node struct {
	id  uuid.UUID
	mux sync.Mutex
	cfg Config

	inputs  map[uuid.UUID]common.Input
	outputs map[uuid.UUID]common.Output
	sources map[uuid.UUID]Source
	sinks   map[uuid.UUID]Sink

	connMux sync.Mutex
	conn    Conn
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
	Id           uuid.UUID                     `json:"id"`
	Config       json.RawMessage               `json:"config"`
	InputConfig  map[uuid.UUID]json.RawMessage `json:"input_config"`
	OutputConfig map[uuid.UUID]json.RawMessage `json:"output_config"`
	State        json.RawMessage               `json:"state"`
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

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("./%s-driver.json", s.driverName), b, 0644)
}

func newDriverStateHolder(name string) *stateHolder {
	return &stateHolder{driverName: name}
}

func New(cfg Config) (*Node, error) {
	d := &Node{
		id:      cfg.Id,
		cfg:     cfg,
		inputs:  make(map[uuid.UUID]common.Input),
		outputs: make(map[uuid.UUID]common.Output),
		sources: make(map[uuid.UUID]Source),
		sinks:   make(map[uuid.UUID]Sink),
	}

	for name, drv := range sources {
		sh := newDriverStateHolder(name)
		id, err := sh.getId()
		if err != nil {
			return nil, err
		}

		err = drv.Init(id, d, sh) // TODO: stop drivers if one fails to start
		if err != nil {
			return nil, err
		}

		//@@@@
		//go func() {
		//	time.Sleep(2 * time.Second)
		//	drv.Configure(types.SourceConfig{
		//		Config: []byte(`{"capturerType": "dxgi"}`),
		//		Inputs: []types.InputConfig{
		//			{
		//				InputId: uuid.MustParse("317cef78-6a94-4ab2-8ccc-8b643bb250f9"),
		//				Config:  []byte(`{"framerate": 240}`),
		//				Outputs: []types.OutputConfig{
		//					{
		//						OutputId: uuid.UUID{0x22},
		//						NodeId:   uuid.UUID{0x33},
		//						SinkId:   uuid.UUID{0x44},
		//						Leds:     8,
		//						Config:   []byte(`{"width": 2560, "height": 2560, "top": 0, "left": 0, "orientation": "landscape"}`),
		//					},
		//				},
		//			},
		//		},
		//	})
		//}()

		d.sources[id] = drv
	}

	for name, drv := range sinks {
		sh := newDriverStateHolder(name)
		id, err := sh.getId()
		if err != nil {
			return nil, err
		}

		err = drv.Start(id, d, sh) // TODO: stop drivers if one fails to start
		if err != nil {
			return nil, err
		}

		d.sinks[id] = drv
	}

	return d, nil
}

func (n *Node) AddInput(in common.Input) {
	fmt.Println("ADD INPUT CALLED", in)

	n.inputs[in.Id()] = in
	//n.inputCfgs[in.OutputId()] = inputConfig{}

	n.connMux.Lock()
	conn := n.conn
	n.connMux.Unlock()

	if conn != nil {
		err := n.write(conn, event.InputConnected{
			Id:       in.Id(),
			SourceId: in.SourceId(),
			Schema:   in.Schema(),
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	go func() {
		// forward messages from input to the network
		for e := range in.Events() {

			// TODO: synchronized render won't work for same-node i/o.
			//  possible solution: instead of sending data to registry, send
			//  directly to sink node, and calculate and send to registry
			//  the RTT from source to sink. the registry can then calculate
			//  how much the sink should offset its render time to match the
			//  latency of the slowest node on the network.
			//if e.SinkId == n.id {
			//	// deliver to local node outputs
			//
			//	for _, out := range e.Outputs {
			//		if _, ok := n.outputs[out.OutputId]; !ok {
			//			fmt.Println("output not found", out.OutputId)
			//			continue
			//		}
			//
			//		n.outputs[out.OutputId].Render(out.Pix)
			//	}
			//
			//	continue
			//}

			e := e
			go func() {
				n.connMux.Lock()
				conn := n.conn
				n.connMux.Unlock()

				if conn == nil {
					return
				}

				var outputs []event.DataOutput
				for _, output := range e.Outputs {
					outputs = append(outputs, event.DataOutput{
						OutputId: output.OutputId,
						SinkId:   output.SinkId,
						Pix:      output.Pix,
					})
				}

				err := n.write(conn, event.Data{
					NodeId:  e.NodeId,
					Outputs: outputs,
					Latency: e.Latency,
				})
				if err != nil {
					fmt.Println("write error:", err)
				}
			}()

			//n.messages <- Message{
			//	Addr: nil, // TODO: registry addr
			//	Event: event.Data{
			//		Event:     event.Event{Type: event.Data},
			//		SessionId: n.sessionId,
			//		Outputs:   outputs,
			//	},
			//}
		}
	}()
}

type Source interface {
	Id() uuid.UUID
	Init(id uuid.UUID, reg common.IORegistry, stateHolder common.StateHolder) error
	Configure(cfg []byte) error
	Schema() []byte
	Close() error
}

type Sink interface {
	Id() uuid.UUID
	Start(id uuid.UUID, reg common.IORegistry, stateHolder common.StateHolder) error
	Configure(cfg []byte) error
	Schema() []byte
	Render(pixs map[uuid.UUID][]color.NRGBA) error
	Stop() error
}

var sources = map[string]Source{}
var sinks = map[string]Sink{}
var driversMux sync.Mutex

func RegisterSource(name string, driver Source) {
	driversMux.Lock()
	defer driversMux.Unlock()

	sources[name] = driver
}

func RegisterSink(name string, driver Sink) {
	driversMux.Lock()
	defer driversMux.Unlock()

	sinks[name] = driver
}

func (n *Node) RemoveInput(id uuid.UUID) {
	//fmt.Println("RemoveInput CALLED", id)

	n.connMux.Lock()
	conn := n.conn
	n.connMux.Unlock()

	if conn != nil {
		err := n.write(conn, event.InputDisconnected{
			Id:       id,
			SourceId: n.inputs[id].SourceId(),
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	delete(n.inputs, id)
}

func (n *Node) AddOutput(out common.Output) {
	//fmt.Println("ADD OUTPUT CALLED", out)

	n.outputs[out.Id()] = out

	n.connMux.Lock()
	conn := n.conn
	n.connMux.Unlock()

	if conn != nil {
		err := n.write(conn, event.OutputConnected{
			Id:     out.Id(),
			SinkId: out.SinkId(),
			Leds:   out.Leds(),
			Schema: out.Schema(),
		})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (n *Node) RemoveOutput(id uuid.UUID) {
	//fmt.Println("REMOVE OUTPUT CALLED", id)
	n.connMux.Lock()
	conn := n.conn
	n.connMux.Unlock()

	if conn != nil {
		err := n.write(conn, event.OutputDisconnected{
			Id:     id,
			SinkId: n.outputs[id].SinkId(),
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	delete(n.outputs, id)
}

func (n *Node) handleConfigureSink(c Conn, e event.ConfigureSink) error {
	sink, ok := n.sinks[e.SinkId]
	if !ok {
		return errors.New("sink not found")
	}

	err := sink.Configure(e.SinkConfig)
	if err != nil {
		return err
	}

	fmt.Println("sink config set", e.SinkId)
	return nil
}
