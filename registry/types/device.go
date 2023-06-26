package types

import (
	"net"

	"github.com/google/uuid"
)

type Device struct {
	id      uuid.UUID
	online  bool
	address net.Addr
	config  *Config
	state   State
}

func (d *Device) Id() uuid.UUID {
	return d.id
}

func (d *Device) SetState(state State) {
	d.state = state
}

func (d *Device) Online() bool {
	return d.online
}

func (d *Device) Config() *Config {
	return d.config
}

func (d *Device) State() State {
	return d.state
}

func NewDevice(address net.Addr) *Device {
	return &Device{
		id:      uuid.New(),
		online:  true,
		address: address,
		config:  nil,
		state:   StateUnknown,
	}
}
