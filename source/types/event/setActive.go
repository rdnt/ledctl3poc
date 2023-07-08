package event

//type Visualizer string
//
//const (
//	VisualizerNone   Visualizer = "none"
//	VisualizerScreen Visualizer = "screen"
//	VisualizerAudio  Visualizer = "audio"
//)
//
//type SetActiveEvent struct {
//	Event
//	SessionId  string            `json:"sessionId"`
//	Visualizer Visualizer        `json:"visualizer"`
//	Sinks      map[string][]Sink `json:"sinks"`
//}
//
//type Sink struct {
//	Id          string    `json:"id"`
//	Address     string    `json:"address"`
//	Leds        int       `json:"leds"`
//	Calibration []float64 `json:"calibration"`
//}
//
//func (e SetActiveEvent) Type() event.Type {
//	return SetActive
//}
