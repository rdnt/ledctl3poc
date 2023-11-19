package _source_old

import (
	"fmt"
	"image/color"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ledctl3/pkg/uuid"

	"ledctl3/_source-old/types/event"

	"ledctl3/_source-old/types"
)

type source struct {
	id   string
	evts chan UpdateEvent
}

func (s *source) Id() string {
	return s.id
}

func (s *source) Start() error {
	fmt.Println("start", s.id)

	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			s.evts <- UpdateEvent{
				Pix:     make([]color.Color, 4*100),
				Latency: 16 * time.Millisecond,
			}
			fmt.Println("produced event", s.id)

			time.Sleep(800 * time.Millisecond)
		}
	}()

	return nil
}

func (s *source) Events() chan UpdateEvent {
	return s.evts
}

func (s *source) Stop() error {
	fmt.Println("stop", s.id)
	return nil
}

func TestHandleSetActiveIdleEvents(t *testing.T) {
	sessId := uuid.NewString()

	e := event.SetActiveEvent{
		Event:      event.SetActiveEvent{}.Type(),
		SessionId:  sessId,
		Visualizer: event.VisualizerAudio,
		Sinks: map[string][]event.Sink{
			"1": {{
				Id: uuid.NewString(),
				Address: (&net.TCPAddr{
					IP:   net.IPv4(192, 168, 1, 11),
					Port: 1234,
				}).String(),
				Leds:        100,
				Calibration: []float64{},
			}},
			"2": {
				{
					Id: uuid.NewString(),
					Address: (&net.TCPAddr{
						IP:   net.IPv4(192, 168, 1, 12),
						Port: 1234,
					}).String(),
					Leds:        100,
					Calibration: []float64{},
				},
				{
					Id: uuid.NewString(),
					Address: (&net.TCPAddr{
						IP:   net.IPv4(192, 168, 1, 13),
						Port: 1234,
					}).String(),
					Leds:        100,
					Calibration: []float64{},
				},
			},
		},
	}

	sources := make(map[string]Input)

	sources["1"] = &source{id: "1", evts: make(chan UpdateEvent)}
	sources["2"] = &source{id: "2", evts: make(chan UpdateEvent)}

	d := New(nil)
	d.ProcessEvent(e)

	assert.Equal(t, d.state, types.StateActive)
	assert.Equal(t, d.sessionId, sessId)

	update1 := <-sources["1"].Events()
	update2 := <-sources["2"].Events()

	assert.Len(t, update1.Pix, 4*100)
	assert.Len(t, update2.Pix, 4*100)

	assert.Equal(t, update1.Latency, 16*time.Millisecond)
	assert.Equal(t, update2.Latency, 16*time.Millisecond)

	e2 := event.SetIdleEvent{
		Event: event.SetIdleEvent{}.Type(),
	}

	d.ProcessEvent(e2)

	assert.Equal(t, d.state, types.StateIdle)
	assert.Equal(t, d.sessionId, "")
}
