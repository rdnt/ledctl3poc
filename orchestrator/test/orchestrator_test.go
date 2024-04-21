package orchestrator_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"ledctl3/node/event"
	"ledctl3/orchestrator"
)

type mockStateHolder struct{}

func (m mockStateHolder) SetState(state orchestrator.State) error {
	return nil
}

func (m mockStateHolder) GetState() (orchestrator.State, error) {
	return orchestrator.State{}, nil
}

type mockConn struct {
	addr string
}

func (c mockConn) Read() ([]byte, error) {
	return nil, nil
}

func (c mockConn) Ping() error {
	return nil
}

func (c mockConn) Addr() string {
	return c.addr
}

func (c mockConn) Latency() time.Duration {
	return 0
}

func (c mockConn) Write(b []byte) error {
	return nil
}

func TestClientConnect(t *testing.T) {
	o := orchestrator.New(mockStateHolder{})

	conn := mockConn{addr: uuid.New().String()}

	t.Run("client connected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("noop if client already connected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
		assert.ErrorContains(t, err, "client already connected")
	})

	t.Run("client disconnected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.DisconnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("noop if client already disconnected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.DisconnectedEvent{})
		assert.ErrorContains(t, err, "client already disconnected")
	})
}

func TestAuthenticate(t *testing.T) {
	o := orchestrator.New(mockStateHolder{})

	conn := mockConn{addr: uuid.New().String()}
	nodeId := uuid.New()

	t.Run("client connected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("auth failure", func(t *testing.T) {
		err := o.ProcessEvent(conn, event.Authenticate{})
		assert.ErrorContains(t, err, "authentication failed")
	})

	t.Run("authenticated", func(t *testing.T) {
		err := o.ProcessEvent(conn, event.Authenticate{
			Id: nodeId,
		})
		assert.NilError(t, err)
	})

	t.Run("noop if already authenticated", func(t *testing.T) {
		err := o.ProcessEvent(conn, event.Authenticate{
			Id: nodeId,
		})
		assert.NilError(t, err)
	})
}

func TestInputConnectedDisconnected(t *testing.T) {
	t.Run("error if not connected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("input connected error", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.Error(t, err, "not connected")
			assert.Equal(t, len(o.State.Nodes), 0)
		})
	})

	t.Run("error if not authenticated", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("input connected error", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.Error(t, err, "not authenticated")
			assert.Equal(t, len(o.State.Nodes), 0)
		})
	})

	t.Run("input connected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("input connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Inputs[inId].Connected, true)
		})
	})

	t.Run("input reconnected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("input connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Inputs[inId].Connected, true)
		})

		t.Run("input disconnected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputDisconnected{
				Id: inId,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Inputs[inId].Connected, false)
		})

		t.Run("input connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Inputs[inId].Connected, true)
		})
	})
}

func TestOutputConnectedDisconnected(t *testing.T) {
	t.Run("error if not connected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("output connected error", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.Error(t, err, "not connected")
			assert.Equal(t, len(o.State.Nodes), 0)
		})
	})

	t.Run("error if not authenticated", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("output connected error", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.Error(t, err, "not authenticated")
			assert.Equal(t, len(o.State.Nodes), 0)
		})
	})

	t.Run("output connected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("out connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Outputs[inId].Connected, true)
		})
	})

	t.Run("output reconnected", func(t *testing.T) {
		o := orchestrator.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("output connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Outputs[inId].Connected, true)
		})

		t.Run("output disconnected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputDisconnected{
				Id: inId,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Outputs[inId].Connected, false)
		})

		t.Run("output connected", func(t *testing.T) {
			err := o.ProcessEvent(conn, event.OutputConnected{
				Id:     inId,
				Schema: nil,
				Config: nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
			assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
			assert.Equal(t, o.State.Nodes[nodeId].Outputs[inId].Connected, true)
		})
	})
}

func TestCreateProfile(t *testing.T) {
	o := orchestrator.New(mockStateHolder{})
	conn := mockConn{addr: uuid.New().String()}
	nodeId := uuid.New()
	inId := uuid.New()
	outId := uuid.New()

	t.Run("node with input and output connected", func(t *testing.T) {
		t.Run("authenticated", func(t *testing.T) {
			err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
			assert.NilError(t, err)

			err = o.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(o.State.Nodes), 1)
		})

		err := o.ProcessEvent(conn, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)
		assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)

		err = o.ProcessEvent(conn, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)

		assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
	})

	t.Run("error if no links", func(t *testing.T) {
		name := "test"
		_, err := o.CreateProfile(name, nil)
		assert.ErrorContains(t, err, "no links specified")
	})

	t.Run("profile created", func(t *testing.T) {
		name := "test"
		link := []orchestrator.Link{
			{
				InputId:  inId,
				OutputId: outId,
				Config:   nil,
			},
		}
		prof, err := o.CreateProfile(name, link)
		assert.NilError(t, err)

		assert.Equal(t, len(o.State.Profiles), 1)
		assert.Equal(t, prof.Name, name)
		assert.DeepEqual(t, prof.Links, link)
	})
}

func TestEnableProfile(t *testing.T) {
	//sh := mockStateHolder{}
	//msgs := make([]message, 0)
	//reg := registry.New(sh, func(addr string, e event.Event) error {
	//	msgs = append(msgs, message{
	//		addr: addr,
	//		e:    e,
	//	})
	//	return nil
	//})
	//
	//addr := uuid.New().String()
	//devId := uuid.New()
	//inId := uuid.New()
	//outId := uuid.New()

	o := orchestrator.New(mockStateHolder{})
	conn := mockConn{addr: uuid.New().String()}
	nodeId := uuid.New()
	inId := uuid.New()
	outId := uuid.New()

	t.Run("cannot enable non-existent profile", func(t *testing.T) {
		err := o.EnableProfile(uuid.New())
		assert.Error(t, err, "profile not found")
	})

	t.Run("node with input and output connected", func(t *testing.T) {
		err := o.ProcessEvent(conn, orchestrator.ConnectedEvent{})
		assert.NilError(t, err)

		err = o.ProcessEvent(conn, event.Authenticate{
			Id: nodeId,
		})
		assert.NilError(t, err)

		err = o.ProcessEvent(conn, event.InputConnected{
			Id:     inId,
			Schema: nil,
			Config: nil,
		})
		assert.NilError(t, err)

		err = o.ProcessEvent(conn, event.OutputConnected{
			Id:     outId,
			Leds:   40,
			Config: nil,
			Schema: nil,
		})
		assert.NilError(t, err)

		assert.Equal(t, len(o.State.Nodes), 1)
		assert.Equal(t, len(o.State.Nodes[nodeId].Inputs), 1)
		assert.Equal(t, len(o.State.Nodes[nodeId].Outputs), 1)
	})

	var id uuid.UUID
	t.Run("profile created", func(t *testing.T) {
		name := "test"
		link := []orchestrator.Link{
			{
				InputId:  inId,
				OutputId: outId,
				Config:   nil,
			},
		}
		prof, err := o.CreateProfile(name, link)
		assert.NilError(t, err)
		id = prof.Id

		assert.Equal(t, len(o.State.Profiles), 1)
		assert.Equal(t, prof.Name, name)
		assert.DeepEqual(t, prof.Links, link)
	})

	t.Run("profile enabled", func(t *testing.T) {
		err := o.EnableProfile(id)
		assert.NilError(t, err)
	})

	//t.Run("enableInput event sent", func(t *testing.T) {
	//	assert.Equal(t, len(msgs), 1)
	//	assert.Equal(t, msgs[0].addr, addr)
	//	assert.DeepEqual(t, msgs[0].e, event.SetInputActive{
	//		Id: inId,
	//		Outputs: []event.SetInputActiveOutput{
	//			{
	//				OutputId: outId,
	//				SinkId:   devId,
	//				Leds:     40,
	//				Config:   nil,
	//			},
	//		},
	//	})
	//})

	t.Run("cannot re-enable profile", func(t *testing.T) {
		err := o.EnableProfile(id)
		assert.Error(t, err, "profile already enabled")
	})

	//t.Run("no additional events sent", func(t *testing.T) {
	//	assert.Equal(t, len(msgs), 1)
	//})
}

//func TestData(t *testing.T) {
//	sh := mockStateHolder{}
//	msgs := make([]message, 0)
//	reg := registry.New(sh, func(addr string, e event.Event) error {
//		fmt.Println(addr, e)
//		msgs = append(msgs, message{
//			addr: addr,
//			e:    e,
//		})
//		return nil
//	})
//
//	addr1 := uuid.New().String()
//	devId1 := uuid.New()
//	inId1 := uuid.New()
//	outId1 := uuid.New()
//
//	addr2 := uuid.New().String()
//	devId2 := uuid.New()
//	inId2 := uuid.New()
//	outId2 := uuid.New()
//
//	t.Run("noop if node disconnected", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.Data{
//			SinkId: devId2,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId2,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.Error(t, err, "node disconnected")
//	})
//
//	t.Run("node 1 connected", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.NodeConnected{Id: devId1})
//		assert.NilError(t, err)
//
//		err = reg.ProcessEvent(addr1, event.InputConnected{
//			Id:     inId1,
//			Schema: nil,
//			Config: nil,
//		})
//		assert.NilError(t, err)
//
//		err = reg.ProcessEvent(addr1, event.OutputConnected{
//			Id:     outId1,
//			Leds:   40,
//			Config: nil,
//			Schema: nil,
//		})
//		assert.NilError(t, err)
//
//		assert.Equal(t, len(reg.State.Nodes), 1)
//		assert.Equal(t, len(reg.State.Nodes[devId1].Inputs), 1)
//		assert.Equal(t, len(reg.State.Nodes[devId1].Outputs), 1)
//	})
//
//	t.Run("node 2 connected", func(t *testing.T) {
//		err := reg.ProcessEvent(addr2, event.NodeConnected{Id: devId2})
//		assert.NilError(t, err)
//
//		err = reg.ProcessEvent(addr2, event.InputConnected{
//			Id:     inId2,
//			Schema: nil,
//			Config: nil,
//		})
//		assert.NilError(t, err)
//
//		err = reg.ProcessEvent(addr2, event.OutputConnected{
//			Id:     outId2,
//			Leds:   80,
//			Config: nil,
//			Schema: nil,
//		})
//		assert.NilError(t, err)
//
//		assert.Equal(t, len(reg.State.Nodes), 2)
//		assert.Equal(t, len(reg.State.Nodes[devId2].Inputs), 1)
//		assert.Equal(t, len(reg.State.Nodes[devId2].Outputs), 1)
//	})
//
//	t.Run("noop if invalid sink", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.Data{
//			SinkId: uuid.New(),
//			Outputs: []event.DataOutput{
//				{
//					OutputId: uuid.New(),
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.Error(t, err, "unknown sink node")
//	})
//
//	t.Run("node disconnected", func(t *testing.T) {
//		reg.HandleDisconnected(addr2)
//	})
//
//	t.Run("noop if invalid sink", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.Data{
//			SinkId: devId2,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId2,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.Error(t, err, "sink node disconnected")
//	})
//
//	t.Run("node connected", func(t *testing.T) {
//		err := reg.ProcessEvent(addr2, event.NodeConnected{Id: devId2})
//		assert.NilError(t, err)
//	})
//
//	t.Run("data events ignored if output inactive", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.Data{
//			SinkId: devId2,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId2,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.Error(t, err, "invalid output")
//
//		err = reg.ProcessEvent(addr2, event.Data{
//			SinkId: devId1,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId1,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.Error(t, err, "invalid output")
//	})
//
//	t.Run("profile enabled", func(t *testing.T) {
//		name := "test"
//		link := []registry.Link{
//			//{
//			//	InputId:  inId1,
//			//	OutputId: outId1,
//			//	Config:   nil,
//			//},
//			{
//				InputId:  inId1,
//				OutputId: outId2,
//				Config:   nil,
//			},
//			{
//				InputId:  inId2,
//				OutputId: outId1,
//				Config:   nil,
//			},
//			//{
//			//	InputId:  inId2,
//			//	OutputId: outId2,
//			//	Config:   nil,
//			//},
//		}
//		prof, err := reg.CreateProfile(name, link)
//		assert.NilError(t, err)
//
//		err = reg.EnableProfile(prof.Id)
//		assert.NilError(t, err)
//	})
//
//	t.Run("data event is processed", func(t *testing.T) {
//		err := reg.ProcessEvent(addr1, event.Data{
//			SinkId: devId2,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId2,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.NilError(t, err)
//
//		err = reg.ProcessEvent(addr2, event.Data{
//			SinkId: devId1,
//			Outputs: []event.DataOutput{
//				{
//					OutputId: outId1,
//					Pix:      make([]color.Color, 40),
//				},
//			},
//		})
//		assert.NilError(t, err)
//	})
//}
