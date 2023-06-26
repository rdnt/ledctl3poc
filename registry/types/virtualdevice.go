package types

import (
	"errors"

	"github.com/google/uuid"
)

type VirtualDevice struct {
	id     uuid.UUID
	devIds []uuid.UUID
	leds   int
	calib  map[int]LedCalibration
	state  State
	online bool
}

func (v *VirtualDevice) Online() bool {
	return v.online
}

func (v *VirtualDevice) Leds() int {
	return v.leds
}

func (v *VirtualDevice) Calibration() map[int]LedCalibration {
	return v.calib
}

func (v *VirtualDevice) State() State {
	return v.state
}

func NewVirtualDevice(ids ...uuid.UUID) (*VirtualDevice, error) {
	if len(ids) == 0 {
		return nil, errors.New("no devices provided")
	}

	return &VirtualDevice{
		id:     uuid.New(),
		devIds: ids,
	}, nil
}
