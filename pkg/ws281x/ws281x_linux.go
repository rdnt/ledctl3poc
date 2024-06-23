package ws281x

import (
	"errors"
	"sync"

	ws281x "github.com/rpi-ws281x/rpi-ws281x-go"
)

// Engine represents a wrapper around the ws281x library.
// Holds the state of the leds, the state of the effects running, the leds count
//
//	and a reference to the underlying ws281x library instance
type Engine struct {
	mux       sync.Mutex
	LedsCount int
	engine    *ws281x.WS2811
	leds      []uint32
}

// Init initializes a new instance of the ws281x library
func Init(gpioPin int, ledCount int, brightness int, stripType string) (*Engine, error) {
	// Initialize ws281x engine
	opt := ws281x.DefaultOptions
	opt.Frequency = 800000
	opt.RenderWaitTime = 0

	opt.Channels[0].Brightness = brightness
	opt.Channels[0].LedCount = ledCount
	opt.Channels[0].GpioPin = gpioPin
	//opt.Channels[0].Gamma = []uint8{
	//	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	//	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29,
	//	30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43,
	//	44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57,
	//	58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71,
	//	72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85,
	//	86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
	//	100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111,
	//	112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123,
	//	124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135,
	//	136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147,
	//	148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159,
	//	160, 161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171,
	//	172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183,
	//	184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195,
	//	196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207,
	//	208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219,
	//	220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231,
	//	232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243,
	//	244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255,
	//}

	st := 0

	switch stripType {
	case "rgbw":
		st = ws281x.SK6812StripRGBW
	case "rbgw":
		st = ws281x.SK6812StripRBGW
	case "grbw":
		st = ws281x.SK6812StripGRBW
	case "gbrw":
		st = ws281x.SK6812StrioGBRW
	case "brgw":
		st = ws281x.SK6812StrioBRGW
	case "bgrw":
		st = ws281x.SK6812StripBGRW
	case "rgb":
		st = ws281x.WS2811StripRGB
	case "rbg":
		st = ws281x.WS2811StripRBG
	case "grb":
		st = ws281x.WS2811StripGRB
	case "gbr":
		st = ws281x.WS2811StripGBR
	case "brg":
		st = ws281x.WS2811StripBRG
	case "bgr":
		st = ws281x.WS2811StripBGR
	default:
		st = ws281x.WS2811StripBGR
	}

	opt.Channels[0].StripeType = st

	ws, err := ws281x.MakeWS2811(&opt)
	if err != nil {
		return nil, err
	}

	err = ws.Init()
	if err != nil {
		return nil, err
	}

	return &Engine{
		LedsCount: ledCount,
		engine:    ws,
		leds:      make([]uint32, ledCount),
	}, nil
}

// Fini does cleanup operations
func (ws *Engine) Fini() {
	ws.engine.Fini()
}

// Clear resets all the leds (turns them off by setting their color to black)
func (ws *Engine) Clear() error {
	ws.leds = make([]uint32, ws.LedsCount)
	return ws.Render()
}

// Render renders the colors saved on the leds array onto the led strip
func (ws *Engine) Render() error {
	leds := ws.leds

	ws.engine.SetLedsSync(0, leds)
	return ws.engine.Render()
}

// SetLedColor changes the color of the led in the specified index
func (ws *Engine) SetLedColor(index int, r uint8, g uint8, b uint8, a uint8) error {
	// WRGB
	color := uint32(a)<<24 | uint32(r)<<16 | uint32(g)<<8 | uint32(b)

	if index >= len(ws.leds) || index < 0 {
		return errors.New("Invalid led index")
	}

	ws.leds[index] = color

	return nil
}
