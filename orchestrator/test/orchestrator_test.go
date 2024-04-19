package orchestrator_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

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
