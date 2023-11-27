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
	return registry.State{}, nil
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
		assert.Error(t, err, "device already disconnected")
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
		assert.Error(t, err, "device already disconnected")
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestInputConnectedDisconnected(t *testing.T) {
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
		assert.Error(t, err, "device disconnected")
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
		assert.Equal(t, reg.State.Devices[devId].Inputs[inId].Connected, true)
	})

	t.Run("input disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputDisconnected{
			Id: inId,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Inputs), 1)
		assert.Equal(t, reg.State.Devices[devId].Inputs[inId].Connected, false)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputDisconnected{
			Id: inId,
		})
		assert.Error(t, err, "device disconnected")
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Inputs), 1)
		assert.Equal(t, reg.State.Devices[devId].Inputs[inId].Connected, false)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestOutputConnectedDisconnected(t *testing.T) {
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
		assert.Error(t, err, "device disconnected")
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

	t.Run("output disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputDisconnected{
			Id: outId,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 1)
		assert.Equal(t, reg.State.Devices[devId].Outputs[outId].Connected, false)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Devices), 1)
	})

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputDisconnected{
			Id: outId,
		})
		assert.Error(t, err, "device disconnected")
		assert.Equal(t, len(reg.State.Devices), 1)
		assert.Equal(t, len(reg.State.Devices[devId].Outputs), 1)
		assert.Equal(t, reg.State.Devices[devId].Outputs[outId].Connected, false)
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

	t.Run("error if no io", func(t *testing.T) {
		name := "test"
		_, err := reg.CreateProfile(name, nil)
		assert.ErrorIs(t, err, registry.ErrEmptyIO)
	})

	t.Run("profile created", func(t *testing.T) {
		name := "test"
		io := []registry.IOConfig{
			{
				InputId:  inId,
				OutputId: outId,
				Config:   nil,
			},
		}
		prof, err := reg.CreateProfile(name, io)
		assert.NilError(t, err)

		assert.Equal(t, len(reg.State.Profiles), 1)
		assert.Equal(t, prof.Name, name)
		assert.DeepEqual(t, prof.IO, io)
	})

	t.Run("no events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 0)
	})
}

func TestEnableProfile(t *testing.T) {
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

	t.Run("cannot enable non-existent profile", func(t *testing.T) {
		err := reg.EnableProfile(uuid.New())
		assert.Error(t, err, "profile not found")
	})

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

	var id uuid.UUID
	t.Run("profile created", func(t *testing.T) {
		name := "test"
		io := []registry.IOConfig{
			{
				InputId:  inId,
				OutputId: outId,
				Config:   nil,
			},
		}
		prof, err := reg.CreateProfile(name, io)
		assert.NilError(t, err)
		id = prof.Id

		assert.Equal(t, len(reg.State.Profiles), 1)
		assert.Equal(t, prof.Name, name)
		assert.DeepEqual(t, prof.IO, io)
	})

	t.Run("profile enabled", func(t *testing.T) {
		err := reg.EnableProfile(id)
		assert.NilError(t, err)
	})

	t.Run("enableInput event sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 1)
		assert.Equal(t, msgs[0].addr, addr)
		assert.DeepEqual(t, msgs[0].e, event.SetInputActive{
			Id: inId,
			Outputs: []event.SetInputActiveOutput{
				{
					Id:     outId,
					Leds:   40,
					Config: nil,
				},
			},
		})
	})

	t.Run("cannot re-enable profile", func(t *testing.T) {
		err := reg.EnableProfile(id)
		assert.Error(t, err, "profile already enabled")
	})

	t.Run("no additional events sent", func(t *testing.T) {
		assert.Equal(t, len(msgs), 1)
	})
}
