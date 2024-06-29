package registry_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	"ledctl3/node/event"
	"ledctl3/registry"
	"ledctl3/registry/types"
)

type mockStateHolder struct{}

func (m mockStateHolder) SetState(state *registry.State) error {
	return nil
}

func (m mockStateHolder) GetState() (*registry.State, error) {
	return nil, nil
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
	r := registry.New(mockStateHolder{})

	conn := mockConn{addr: uuid.New().String()}

	t.Run("client connected", func(t *testing.T) {
		err := r.ProcessEvent(conn, registry.ConnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("noop if client already connected", func(t *testing.T) {
		err := r.ProcessEvent(conn, registry.ConnectedEvent{})
		assert.ErrorContains(t, err, "client already connected")
	})

	t.Run("client disconnected", func(t *testing.T) {
		err := r.ProcessEvent(conn, registry.DisconnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("noop if client already disconnected", func(t *testing.T) {
		err := r.ProcessEvent(conn, registry.DisconnectedEvent{})
		assert.ErrorContains(t, err, "client already disconnected")
	})
}

func TestAuthenticate(t *testing.T) {
	r := registry.New(mockStateHolder{})

	conn := mockConn{addr: uuid.New().String()}
	nodeId := uuid.New()

	t.Run("client connected", func(t *testing.T) {
		err := r.ProcessEvent(conn, registry.ConnectedEvent{})
		assert.NilError(t, err)
	})

	t.Run("auth failure", func(t *testing.T) {
		err := r.ProcessEvent(conn, event.Authenticate{})
		assert.ErrorContains(t, err, "authentication failed")
	})

	t.Run("authenticated", func(t *testing.T) {
		err := r.ProcessEvent(conn, event.Authenticate{
			Id: nodeId,
		})
		assert.NilError(t, err)
	})

	t.Run("noop if already authenticated", func(t *testing.T) {
		err := r.ProcessEvent(conn, event.Authenticate{
			Id: nodeId,
		})
		assert.NilError(t, err)
	})
}

func TestInputConnectedDisconnected(t *testing.T) {
	t.Run("error if not connected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("input connected error", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
			})
			assert.Error(t, err, "not connected")
			assert.Equal(t, len(r.State.Nodes), 0)
		})
	})

	t.Run("error if not authenticated", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("input connected error", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:     inId,
				Schema: nil,
			})
			assert.Error(t, err, "not authenticated")
			assert.Equal(t, len(r.State.Nodes), 0)
		})
	})

	t.Run("error if source is unknown", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		srcId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("got unknown source", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:       inId,
				SourceId: srcId,
				Schema:   nil,
			})
			assert.ErrorContains(t, err, "unknown source")
		})
	})

	t.Run("input connected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		srcId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("capabilities received", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Capabilities{
				Sources: []event.CapabilitiesSource{
					{
						Id:     srcId,
						Schema: nil,
					},
				},
			})
			assert.NilError(t, err)
		})

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:       inId,
				SourceId: srcId,
				Schema:   nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SourceIds), 1)
			assert.Equal(t, len(r.State.Sources[srcId].InputIds), 1)
			assert.Equal(t, r.State.Inputs[inId].Connected, true)
		})
	})

	t.Run("input reconnected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		srcId := uuid.New()
		inId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("capabilities received", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Capabilities{
				Sources: []event.CapabilitiesSource{
					{
						Id:     srcId,
						Schema: nil,
					},
				},
			})
			assert.NilError(t, err)
		})

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:       inId,
				SourceId: srcId,
				Schema:   nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SourceIds), 1)
			assert.Equal(t, len(r.State.Sources[srcId].InputIds), 1)
			assert.Equal(t, r.State.Inputs[inId].Connected, true)
		})

		t.Run("input disconnected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputDisconnected{
				SourceId: srcId,
				Id:       inId,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SourceIds), 1)
			assert.Equal(t, len(r.State.Sources[srcId].InputIds), 1)
			assert.Equal(t, r.State.Inputs[inId].Connected, false)
		})

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:       inId,
				SourceId: srcId,
				Schema:   nil,
			})
			assert.NilError(t, err)
			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SourceIds), 1)
			assert.Equal(t, len(r.State.Sources[srcId].InputIds), 1)
			assert.Equal(t, r.State.Inputs[inId].Connected, true)
		})
	})
}

func TestOutputConnectedDisconnected(t *testing.T) {
	t.Run("error if not connected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		sinkId := uuid.New()
		outId := uuid.New()

		t.Run("output connected error", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.Error(t, err, "not connected")
			assert.Equal(t, len(r.State.Nodes), 0)
			assert.Equal(t, len(r.State.Sinks), 0)
			assert.Equal(t, len(r.State.Outputs), 0)
		})
	})

	t.Run("error if not authenticated", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		sinkId := uuid.New()
		outId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("output connected error", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.Error(t, err, "not authenticated")
			assert.Equal(t, len(r.State.Nodes), 0)
			assert.Equal(t, len(r.State.Sinks), 0)
			assert.Equal(t, len(r.State.Outputs), 0)
		})
	})

	t.Run("error if sink is unknown", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		sinkId := uuid.New()
		outId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
		})

		t.Run("got unknown sink", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.ErrorContains(t, err, "unknown sink")

			assert.Equal(t, len(r.State.Sinks), 0)
			assert.Equal(t, len(r.State.Outputs), 0)
		})
	})

	t.Run("output connected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		sinkId := uuid.New()
		outId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
		})

		t.Run("capabilities received", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Capabilities{
				Sinks: []event.CapabilitiesSink{
					{
						Id:     sinkId,
						Schema: nil,
					},
				},
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Sinks), 1)
		})

		t.Run("output connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SinkIds), 1)
			assert.Equal(t, len(r.State.Sinks[sinkId].OutputIds), 1)
			assert.Equal(t, r.State.Outputs[outId].Connected, true)
		})
	})

	t.Run("input reconnected", func(t *testing.T) {
		r := registry.New(mockStateHolder{})
		conn := mockConn{addr: uuid.New().String()}
		nodeId := uuid.New()
		sinkId := uuid.New()
		outId := uuid.New()

		t.Run("node connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)
		})

		t.Run("node authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		t.Run("capabilities received", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Capabilities{
				Sinks: []event.CapabilitiesSink{
					{
						Id:     sinkId,
						Schema: nil,
					},
				},
			})
			assert.NilError(t, err)
		})

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SinkIds), 1)
			assert.Equal(t, len(r.State.Sinks[sinkId].OutputIds), 1)
			assert.Equal(t, r.State.Outputs[outId].Connected, true)
		})

		t.Run("input disconnected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputDisconnected{
				SinkId: sinkId,
				Id:     outId,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SinkIds), 1)
			assert.Equal(t, len(r.State.Sinks[sinkId].OutputIds), 1)
			assert.Equal(t, r.State.Outputs[outId].Connected, false)
		})

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Nodes), 1)
			assert.Equal(t, len(r.State.Nodes[nodeId].SinkIds), 1)
			assert.Equal(t, len(r.State.Sinks[sinkId].OutputIds), 1)
			assert.Equal(t, r.State.Outputs[outId].Connected, true)
		})
	})
}

func TestCreateEnableProfile(t *testing.T) {
	r := registry.New(mockStateHolder{})
	conn := mockConn{addr: uuid.New().String()}
	nodeId := uuid.New()
	srcId := uuid.New()
	sinkId := uuid.New()
	inId := uuid.New()
	outId := uuid.New()

	t.Run("node with input and output connected", func(t *testing.T) {
		t.Run("authenticated", func(t *testing.T) {
			err := r.ProcessEvent(conn, registry.ConnectedEvent{})
			assert.NilError(t, err)

			err = r.ProcessEvent(conn, event.Authenticate{
				Id: nodeId,
			})
			assert.NilError(t, err)
		})

		assert.Equal(t, len(r.State.Nodes), 1)

		t.Run("capabilities received", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.Capabilities{
				Sources: []event.CapabilitiesSource{
					{
						Id:     srcId,
						Schema: nil,
					},
				},
				Sinks: []event.CapabilitiesSink{
					{
						Id:     sinkId,
						Schema: nil,
					},
				},
			})
			assert.NilError(t, err)
		})

		assert.Equal(t, len(r.State.Nodes[nodeId].SourceIds), 1)
		assert.Equal(t, len(r.State.Nodes[nodeId].SinkIds), 1)
		assert.Equal(t, len(r.State.Sources), 1)
		assert.Equal(t, len(r.State.Sinks), 1)

		t.Run("input connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.InputConnected{
				Id:       inId,
				SourceId: srcId,
				Schema:   nil,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Sources[srcId].InputIds), 1)
			assert.Equal(t, len(r.State.Inputs), 1)

		})

		t.Run("output connected", func(t *testing.T) {
			err := r.ProcessEvent(conn, event.OutputConnected{
				Id:     outId,
				SinkId: sinkId,
				Leds:   40,
				Schema: nil,
			})
			assert.NilError(t, err)

			assert.Equal(t, len(r.State.Sinks[sinkId].OutputIds), 1)
			assert.Equal(t, len(r.State.Outputs), 1)
		})

	})

	profId := uuid.New()

	t.Run("profile created", func(t *testing.T) {
		name := "test"
		links := []*types.Link{
			{
				InputId:     inId,
				OutputId:    outId,
				InputConfig: nil,
			},
		}
		err := r.CreateProfile(profId, name, links)
		assert.NilError(t, err)

		assert.Equal(t, len(r.State.Profiles), 1)
		assert.Equal(t, r.State.Profiles[profId].Name, name)
		linkId := r.State.Profiles[profId].LinkIds[0]
		assert.DeepEqual(t, r.State.Links[linkId].InputId, links[0].InputId)
		assert.DeepEqual(t, r.State.Links[linkId].OutputId, links[0].OutputId)
		assert.DeepEqual(t, r.State.Links[linkId].InputConfig, links[0].InputConfig)
	})

	t.Run("profile enabled", func(t *testing.T) {
		err := r.EnableProfile(profId)
		assert.NilError(t, err)
	})

	t.Run("cannot re-enable profile", func(t *testing.T) {
		err := r.EnableProfile(profId)
		assert.Error(t, err, "profile already enabled")
	})
}
