package registry_test

import (
	"fmt"
	"image/color"
	"testing"

	"gotest.tools/v3/assert"

	"ledctl3/node/event"
	"ledctl3/pkg/uuid"
	"ledctl3/registry"
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

	t.Run("node connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("noop if connect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: id})
		assert.Error(t, err, "node already connected")
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("node disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("node reconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
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

	t.Run("unknown node disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.Error(t, err, "device already disconnected")
		assert.Equal(t, len(reg.State.Nodes), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: id})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("noop if disconnect is sent again", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.Error(t, err, "device already disconnected")
		assert.Equal(t, len(reg.State.Nodes), 1)
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
		assert.Equal(t, len(reg.State.Nodes), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: devId})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 0)
	})

	t.Run("input connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 1)
		assert.Equal(t, reg.State.Nodes[devId].Inputs[inId].Connected, true)
	})

	t.Run("input disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputDisconnected{
			Id: inId,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 1)
		assert.Equal(t, reg.State.Nodes[devId].Inputs[inId].Connected, false)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.InputDisconnected{
			Id: inId,
		})
		assert.Error(t, err, "device disconnected")
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 1)
		assert.Equal(t, reg.State.Nodes[devId].Inputs[inId].Connected, false)
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
		assert.Equal(t, len(reg.State.Nodes), 0)
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: devId})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 0)
	})

	t.Run("output connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 1)
	})

	t.Run("output disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputDisconnected{
			Id: outId,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 1)
		assert.Equal(t, reg.State.Nodes[devId].Outputs[outId].Connected, false)
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, registry.DisconnectedEvent{})
		assert.NilError(t, err)
		assert.Equal(t, len(reg.State.Nodes), 1)
	})

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr, event.OutputDisconnected{
			Id: outId,
		})
		assert.Error(t, err, "device disconnected")
		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 1)
		assert.Equal(t, reg.State.Nodes[devId].Outputs[outId].Connected, false)
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
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: devId})
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

		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 1)
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
		err := reg.ProcessEvent(addr, event.NodeConnected{Id: devId})
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

		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Inputs), 1)
		assert.Equal(t, len(reg.State.Nodes[devId].Outputs), 1)
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
					OutputId: outId,
					SinkId:   devId,
					Leds:     40,
					Config:   nil,
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

func TestData(t *testing.T) {
	sh := mockStateHolder{}
	msgs := make([]message, 0)
	reg := registry.New(sh, func(addr string, e event.Event) error {
		fmt.Println(addr, e)
		msgs = append(msgs, message{
			addr: addr,
			e:    e,
		})
		return nil
	})

	addr1 := uuid.New().String()
	devId1 := uuid.New()
	inId1 := uuid.New()
	outId1 := uuid.New()

	addr2 := uuid.New().String()
	devId2 := uuid.New()
	inId2 := uuid.New()
	outId2 := uuid.New()

	t.Run("noop if device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.Data{
			SinkId: devId2,
			Outputs: []event.DataOutput{
				{
					OutputId: outId2,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.Error(t, err, "device disconnected")
	})

	t.Run("device 1 connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.NodeConnected{Id: devId1})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr1, event.InputConnected{
			Id:     inId1,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr1, event.OutputConnected{
			Id:     outId1,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)

		assert.Equal(t, len(reg.State.Nodes), 1)
		assert.Equal(t, len(reg.State.Nodes[devId1].Inputs), 1)
		assert.Equal(t, len(reg.State.Nodes[devId1].Outputs), 1)
	})

	t.Run("device 2 connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr2, event.NodeConnected{Id: devId2})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr2, event.InputConnected{
			Id:     inId2,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr2, event.OutputConnected{
			Id:     outId2,
			Leds:   80,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)

		assert.Equal(t, len(reg.State.Nodes), 2)
		assert.Equal(t, len(reg.State.Nodes[devId2].Inputs), 1)
		assert.Equal(t, len(reg.State.Nodes[devId2].Outputs), 1)
	})

	t.Run("noop if invalid sink", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.Data{
			SinkId: uuid.New(),
			Outputs: []event.DataOutput{
				{
					OutputId: uuid.New(),
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.Error(t, err, "unknown sink device")
	})

	t.Run("device disconnected", func(t *testing.T) {
		err := reg.ProcessEvent(addr2, registry.DisconnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("noop if invalid sink", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.Data{
			SinkId: devId2,
			Outputs: []event.DataOutput{
				{
					OutputId: outId2,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.Error(t, err, "sink device disconnected")
	})

	t.Run("device connected", func(t *testing.T) {
		err := reg.ProcessEvent(addr2, event.NodeConnected{Id: devId2})
		assert.NilError(t, err)
	})

	t.Run("data events ignored if output inactive", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.Data{
			SinkId: devId2,
			Outputs: []event.DataOutput{
				{
					OutputId: outId2,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.Error(t, err, "invalid output")

		err = reg.ProcessEvent(addr2, event.Data{
			SinkId: devId1,
			Outputs: []event.DataOutput{
				{
					OutputId: outId1,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.Error(t, err, "invalid output")
	})

	t.Run("profile enabled", func(t *testing.T) {
		name := "test"
		io := []registry.IOConfig{
			//{
			//	InputId:  inId1,
			//	OutputId: outId1,
			//	Config:   nil,
			//},
			{
				InputId:  inId1,
				OutputId: outId2,
				Config:   nil,
			},
			{
				InputId:  inId2,
				OutputId: outId1,
				Config:   nil,
			},
			//{
			//	InputId:  inId2,
			//	OutputId: outId2,
			//	Config:   nil,
			//},
		}
		prof, err := reg.CreateProfile(name, io)
		assert.NilError(t, err)

		err = reg.EnableProfile(prof.Id)
		assert.NilError(t, err)
	})

	t.Run("data event is processed", func(t *testing.T) {
		err := reg.ProcessEvent(addr1, event.Data{
			SinkId: devId2,
			Outputs: []event.DataOutput{
				{
					OutputId: outId2,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.NilError(t, err)

		err = reg.ProcessEvent(addr2, event.Data{
			SinkId: devId1,
			Outputs: []event.DataOutput{
				{
					OutputId: outId1,
					Pix:      make([]color.Color, 40),
				},
			},
		})
		assert.NilError(t, err)
	})
}
