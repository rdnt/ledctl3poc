package event

type Visualizer string

const (
	VisualizerNone   Visualizer = "none"
	VisualizerScreen Visualizer = "screen"
	VisualizerAudio  Visualizer = "audio"
)

type SetActiveEvent struct {
	Event      Type       `json:"event"`
	SessionId  string     `json:"sessionId"`
	Leds       int        `json:"leds"`
	Visualizer Visualizer `json:"visualizer"`
}

func (e SetActiveEvent) Type() Type {
	return SetActive
}
