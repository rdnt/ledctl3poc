package registry_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"ledctl3/event"
	"ledctl3/internal/registry"
	"ledctl3/pkg/uuid"
)

type mockStateHolder struct{}

func (m mockStateHolder) SetState(state registry.State) error {
	return nil
}

func (m mockStateHolder) GetState() (registry.State, error) {
	return registry.State{
		Devices:  make(map[uuid.UUID]*registry.Device),
		Profiles: make(map[uuid.UUID]registry.Profile),
	}, nil
}

type message struct {
	addr string
	e    event.Event
}

func TestConnect(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr := uuid.New().String()
	id := uuid.New()

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("noop if connect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.ErrorIs(t, err, registry.ErrDeviceConnected)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("device reconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestDisconnect(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr := uuid.New().String()
	id := uuid.New()

	t.Run("unknown device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.ErrorIs(t, err, registry.ErrDeviceDisconnected)
		assert.Equal(t, len(reg.State.Devices), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("noop if disconnect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.ErrorIs(t, err, registry.ErrDeviceDisconnected)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestInputConnected(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr := uuid.New().String()
	devId := uuid.New()
	inId := uuid.New()

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.ErrorIs(t, err, registry.ErrDeviceDisconnected)
		assert.Equal(t, len(reg.State.Devices), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: devId})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Inputs), 0)
	})

	t.Run("input connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Inputs), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestOutputConnected(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr := uuid.New().String()
	devId := uuid.New()
	outId := uuid.New()

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.ErrorIs(t, err, registry.ErrDeviceDisconnected)
		assert.Equal(t, len(reg.State.Devices), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: devId})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 0)
	})

	t.Run("output connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestCreateProfile(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr := uuid.New().String()
	devId := uuid.New()
	inId := uuid.New()
	outId := uuid.New()

	t.Run("device with input and output connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: devId})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)

		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Inputs), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 1)
	})

	t.Run("profile created", func(t *testing.T) {
		prof, err := reg.CreateProfile("test", []registry.IOConfig{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}
