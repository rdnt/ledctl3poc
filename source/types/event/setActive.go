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
	Visualizer Visualizer `json:"visualizer"`
	Sources    []Source   `json:"sources"`
	Sinks      []Sink     `json:"sinks"`
}

type Source struct {
	Id string `json:"id"`
}

type Sink struct {
	Id          string    `json:"id"`
	Address     string    `json:"address"`
	Leds        int       `json:"leds"`
	Calibration []float64 `json:"calibration"`
}

func (e SetActiveEvent) Type() Type {
	return SetActive
}
