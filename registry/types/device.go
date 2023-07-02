package types

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

type Device struct {
	id          uuid.UUID
	name        string
	online      bool
	address     net.Addr
	leds        int
	calibration map[int]Calibration
	state       State
}

func (d *Device) Id() uuid.UUID {
	return d.id
}

func (d *Device) Name() string {
	return d.name
}

func (d *Device) Address() net.Addr {
	return d.address
}

func (d *Device) Leds() int {
	return d.leds
}

func (d *Device) SetLeds(leds int) {
	d.leds = leds
}

func (d *Device) Calibration() map[int]Calibration {
	return d.calibration
}

func (d *Device) State() State {
	return d.state
}

func (d *Device) SetState(state State) {
	d.state = state
}

func (d *Device) String() string {
	return fmt.Sprintf(
		"dev{id: %s, name: %s, address: %s, leds: %d, calibration: %v, state: %s}",
		d.id, d.name, d.address, d.leds, d.calibration, d.state,
	)
}

func (d *Device) Handle(e Event) {
	fmt.Printf("device %s: handling event %s\n", d.name, e)
}

func NewDevice(name string, address net.Addr) *Device {
	return &Device{
		id:          uuid.New(),
		name:        name,
		online:      true,
		address:     address,
		leds:        0,
		calibration: make(map[int]Calibration),
		state:       StateOffline,
	}
}
