package registry

import (
	"encoding/json"
	"fmt"

	"ledctl3/pkg/uuid"
)

type Node struct {
	Id        uuid.UUID             `json:"id"`
	Inputs    map[uuid.UUID]*Input  `json:"inputs"`
	Outputs   map[uuid.UUID]*Output `json:"outputs"`
	Drivers   map[uuid.UUID]*Driver `json:"drivers"`
	Connected bool                  `json:"-"`
}

type Driver struct {
	Id     uuid.UUID       `json:"id"`
	Config json.RawMessage `json:"config"`
}

func NewNode(id uuid.UUID, connected bool, drivers map[uuid.UUID]*Driver) *Node {
	return &Node{
		Id:        id,
		Inputs:    make(map[uuid.UUID]*Input),
		Outputs:   make(map[uuid.UUID]*Output),
		Drivers:   drivers,
		Connected: connected,
	}
}

func (d *Node) Disconnect() {
	for _, in := range d.Inputs {
		in.Disconnect()
	}

	for _, out := range d.Outputs {
		out.Disconnect()
	}

	d.Connected = false
	fmt.Println("device disconnected:", d.Id)
}

func (d *Node) ConnectOutput(id, driverId uuid.UUID, leds int, schema, config []byte) {
	out, ok := d.Outputs[id]
	if !ok {
		out = NewOutput(id, driverId, leds, schema, config, true)

		if d.Outputs == nil {
			d.Outputs = make(map[uuid.UUID]*Output)
		}

		d.Outputs[out.Id] = out
	}

	out.Connect()
}

func (d *Node) ConnectInput(id, driverId uuid.UUID, schema, config []byte) {
	in, ok := d.Inputs[id]
	if !ok {
		in = NewInput(id, driverId, schema, config, true)

		if d.Inputs == nil {
			d.Inputs = make(map[uuid.UUID]*Input)
		}

		d.Inputs[in.Id] = in
	}

	in.Connect()
}

func (d *Node) DisconnectInput(id uuid.UUID) {
	in, ok := d.Inputs[id]
	if !ok {
		fmt.Println("input does not exist")
		return
	}

	in.Disconnect()
}

func (d *Node) DisconnectOutput(id uuid.UUID) {
	out, ok := d.Outputs[id]
	if !ok {
		fmt.Println("output does not exist")
		return
	}

	out.Disconnect()
}

func (d *Node) Connect() {
	d.Connected = true
}
