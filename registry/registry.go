package registry

import "ledctl3/registry/types"

type Registry struct {
	devs []Device
	//devs  []*types.Device
	//vdevs []*types.VirtualDevice
}

func New() *Registry {
	return &Registry{
		devs: []Device{},
	}
}

type Device interface {
	Online() bool
	Leds() int
	Calibration() map[int]types.LedCalibration
	State() types.State
}

func (r *Registry) AddDevice(dev Device) {
	r.devs = append(r.devs, dev)
}

func (r *Registry) Devices() []Device {
	return r.devs
}
