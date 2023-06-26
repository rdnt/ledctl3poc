package types

type Config struct {
	Leds        int
	Calibration map[int]LedCalibration
	GroupId     string
}

type LedCalibration struct {
	R float64
	G float64
	B float64
	A float64
}
