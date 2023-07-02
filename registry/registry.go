package registry

import (
	"errors"
	"github.com/google/uuid"
	"ledctl3/registry/types"
	"net"
)

type Registry struct {
	srcs     map[uuid.UUID]Source
	devs     map[uuid.UUID]Device
	profiles map[uuid.UUID]Profile
}

type Profile struct {
	Id       uuid.UUID
	Name     string
	Mappings []map[uuid.UUID]uuid.UUID
}

func New() *Registry {
	return &Registry{
		devs:     map[uuid.UUID]Device{},
		srcs:     map[uuid.UUID]Source{},
		profiles: map[uuid.UUID]Profile{},
	}
}

type Source interface {
	Id() uuid.UUID
	Name() string
	Address() net.Addr
	State() types.State
	SetState(state types.State)
	String() string
	Events() chan types.Event
}

type Device interface {
	Id() uuid.UUID
	Name() string
	Leds() int
	Calibration() map[int]types.Calibration
	State() types.State
	SetState(state types.State)
	String() string
	Handle(e types.Event)
}

var (
	ErrDeviceExists   = errors.New("device already exists")
	ErrDeviceNotFound = errors.New("device not found")
	ErrConfigNotFound = errors.New("config not found")
)

func (r *Registry) AddSource(src *types.Source) error {
	_, ok := r.srcs[src.Id()]
	if ok {
		return ErrDeviceExists
	}

	r.srcs[src.Id()] = src

	return nil
}

func (r *Registry) AddDevice(dev Device) error {
	_, ok := r.devs[dev.Id()]
	if ok {
		return ErrDeviceExists
	}

	if vdev, ok := dev.(*types.VirtualDevice); ok {
		for _, d2 := range vdev.Devices() {
			if _, ok2 := r.devs[d2.Id()]; !ok2 {
				return ErrDeviceNotFound
			}
		}
	}

	r.devs[dev.Id()] = dev

	return nil
}

func (r *Registry) Devices() map[uuid.UUID]Device {
	return r.devs
}

func (r *Registry) AddProfile(name string, mappings []map[uuid.UUID]uuid.UUID) Profile {
	cfg := Profile{
		Id:       uuid.New(),
		Name:     name,
		Mappings: mappings,
	}

	r.profiles[cfg.Id] = cfg
	return cfg
}

func (r *Registry) SelectProfile(id uuid.UUID, state types.State) error {
	cfg, ok := r.profiles[id]
	if !ok {
		return ErrConfigNotFound
	}

	for _, mapping := range cfg.Mappings {
		for srcId, devId := range mapping {
			err := r.setState(srcId, devId, state)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Registry) setState(srcId uuid.UUID, devId uuid.UUID, state types.State) error {
	_, ok := r.srcs[srcId]
	if !ok {
		return ErrDeviceNotFound
	}

	_, ok = r.devs[devId]
	if !ok {
		return ErrDeviceNotFound
	}

	go func() {
		for e := range r.srcs[srcId].Events() {
			r.devs[devId].Handle(e)
		}
	}()

	// prepare the server to start receiving the events
	r.devs[devId].SetState(state)

	// switch state on the source to start event transmission
	r.srcs[srcId].SetState(state)

	return nil
}
