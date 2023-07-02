package source

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"ledctl3/source/types"
	"ledctl3/source/types/event"
	"testing"
)

func TestHandleSetActiveIdleEvents(t *testing.T) {
	sessId := uuid.NewString()

	e := event.SetActiveEvent{
		Event:      event.SetActiveEvent{}.Type(),
		SessionId:  sessId,
		Leds:       100,
		Visualizer: event.VisualizerAudio,
	}

	src := New(nil)
	src.ProcessEvent(e)

	assert.Equal(t, src.state, types.StateActive)
	assert.Equal(t, src.sessionId, sessId)
	assert.Equal(t, src.leds, 100)
	assert.Equal(t, src.visualizer, event.VisualizerAudio)

	e2 := event.SetIdleEvent{
		Event: event.SetIdleEvent{}.Type(),
	}

	src.ProcessEvent(e2)

	assert.Equal(t, src.state, types.StateIdle)
	assert.Equal(t, src.sessionId, "")
	assert.Equal(t, src.visualizer, event.VisualizerNone)
}
