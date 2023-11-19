package registry

import (
	"fmt"

	"ledctl3/pkg/uuid"
)

type Device struct {
	Id        uuid.UUID             `json:"id"`
	Inputs    map[uuid.UUID]*Input  `json:"inputs"`
	Outputs   map[uuid.UUID]*Output `json:"outputs"`
	Connected bool                  `json:"-"`
}

func NewDevice(id uuid.UUID, connected bool) *Device {
	return &Device{
		Id:        id,
		Inputs:    make(map[uuid.UUID]*Input),
		Outputs:   make(map[uuid.UUID]*Output),
		Connected: connected,
	}
}

func (d *Device) Disconnect() {
	for _, in := range d.Inputs {
		in.Disconnect()
	}

	for _, out := range d.Outputs {
		out.Disconnect()
	}

	d.Connected = false
	fmt.Println("device disconnected:", d.Id)
}

func (d *Device) ConnectOutput(id uuid.UUID, leds int) {
	out, ok := d.Outputs[id]
	if !ok {
		out = NewOutput(id, leds, true)
		d.Outputs[out.Id] = out
	}

	out.Connect()
}

func (d *Device) ConnectInput(id uuid.UUID, typ string) {
	in, ok := d.Inputs[id]
	if !ok {
		in = NewInput(id, typ, true)
		d.Inputs[in.Id] = in
	}

	in.Connect()
}

func (d *Device) DisconnectInput(id uuid.UUID) {
	in, ok := d.Inputs[id]
	if !ok {
		fmt.Println("input does not exist")
		return
	}

	in.Disconnect()
}

func (d *Device) DisconnectOutput(id uuid.UUID) {
	out, ok := d.Outputs[id]
	if !ok {
		fmt.Println("output does not exist")
		return
	}

	out.Disconnect()
}

func (d *Device) Connect() {
	d.Connected = true
}
