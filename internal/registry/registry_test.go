package registry

import (
	"testing"

	"gotest.tools/v3/assert"

	"ledctl3/event"
	"ledctl3/pkg/uuid"
)

type mockStateHolder struct{}

func (m mockStateHolder) SetState(state State) error {
	return nil
}

func (m mockStateHolder) GetState() (State, error) {
	return State{
		Devices:  make(map[uuid.UUID]*Device),
		Profiles: make(map[uuid.UUID]Profile),
	}, nil
}

func TestConnect(t *testing.T) {
	sh := mockStateHolder{}
	reg := New(sh, func(addr string, e event.Event) error {
		return nil
	})

	addr := uuid.New().String()
	id := uuid.New()

	t.Run("device connects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.state.Devices), 1)
	})

	t.Run("noop if connect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.ErrorIs(t, err, ErrDeviceConnected)
		assert.Equal(t, len(reg.state.Devices), 1)
	})

	t.Run("device disconnects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.state.Devices), 1)
	})

	t.Run("device reconnects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.state.Devices), 1)
	})
}

func TestDisconnect(t *testing.T) {
	sh := mockStateHolder{}
	reg := New(sh, func(addr string, e event.Event) error {
		return nil
	})

	addr := uuid.New().String()
	id := uuid.New()

	t.Run("unknown device disconnects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.ErrorIs(t, err, ErrDeviceDisconnected)
		assert.Equal(t, len(reg.state.Devices), 0)
	})

	t.Run("device connects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Connect{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.state.Devices), 1)
	})

	t.Run("device disconnects", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.state.Devices), 1)
	})

	t.Run("noop if disconnect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.Disconnect{})
		assert.ErrorIs(t, err, ErrDeviceDisconnected)
		assert.Equal(t, len(reg.state.Devices), 1)
	})
}
