package types

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type VirtualDevice struct {
	id     uuid.UUID
	name   string
	devs   []*Device
	state  State
	online bool
}

func (v *VirtualDevice) Id() uuid.UUID {
	return v.id
}

func (v *VirtualDevice) Name() string {
	return v.name
}

func (v *VirtualDevice) Leds() int {
	var leds int
	for _, dev := range v.devs {
		leds += dev.Leds()
	}

	return leds
}

func (v *VirtualDevice) Calibration() map[int]Calibration {
	calib := make(map[int]Calibration)

	var acc int
	for _, dev := range v.devs {
		for i, c := range dev.Calibration() {
			calib[i+acc] = c
		}

		acc += dev.Leds()
	}

	return calib
}

func (v *VirtualDevice) State() State {
	return v.state
}

func (v *VirtualDevice) String() string {
	return fmt.Sprintf(
		"vdev{id: %s, name: %s, leds: %d, calibration: %v, state: %s}",
		v.id, v.name, v.Leds(), v.Calibration(), v.state,
	)
}

func (v *VirtualDevice) Devices() map[uuid.UUID]*Device {
	return lo.SliceToMap(v.devs, func(dev *Device) (uuid.UUID, *Device) {
		return dev.Id(), dev
	})
}

func (v *VirtualDevice) SetState(state State) {
	v.state = state
}

func (v *VirtualDevice) Handle(e Event) {
	for _, dev := range v.devs {
		// TODO: logic to split a single producer event into multiple consumers
		dev.Handle(e)
	}
}

func NewVirtualDevice(name string, devs ...*Device) (*VirtualDevice, error) {
	if len(devs) == 0 {
		return nil, errors.New("no devices provided")
	}

	return &VirtualDevice{
		id:    uuid.New(),
		name:  name,
		devs:  devs,
		state: StateOffline,
	}, nil
}
