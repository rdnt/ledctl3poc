package audio

import (
	"fmt"
	"image/color"
	"math"

	"github.com/google/uuid"
	gcolor "github.com/gookit/color"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/sgreben/piecewiselinear"

	"ledctl3/pkg/gradient"
	"ledctl3/pkg/pixavg"
	"ledctl3/pkg/sliceavg"
)

type outputProcessor struct {
	outputId uuid.UUID
	nodeId   uuid.UUID
	sinkId   uuid.UUID
	leds     int

	//average holds a pixavg.Average for each segment. The decay rate
	//is affected by windowSize.
	average pixavg.MovingAverage

	averageMagn sliceavg.Average

	gradient gradient.Gradient

	// blackPoint represents the normalization black point as a float value in
	// the range 0-1
	blackPoint float64

	cfg OutputConfig
}

func newOutputProcessor(cfg OutputConfig, windowSize int) *outputProcessor {
	prev := make([]color.Color, cfg.Leds)
	for i := 0; i < len(prev); i++ {
		prev[i] = color.RGBA{}
	}

	// TODO @@@ colors
	grad, err := gradient.New(
		color.NRGBA{R: 0x4a, G: 0x15, B: 0x24, A: 0xff}, // "#4a1524",
		color.NRGBA{R: 0x06, G: 0x53, B: 0x94, A: 0xff}, // "#065394",
		color.NRGBA{R: 0x00, G: 0xb5, B: 0x85, A: 0xff}, // "#00b585",
		color.NRGBA{R: 0xd6, G: 0x60, B: 0xa4, A: 0xff}, // "#d600a4",
		color.NRGBA{R: 0xff, G: 0x00, B: 0x4c, A: 0xff}, // "#ff004c"
	)
	if err != nil {
		panic(err)
	}

	return &outputProcessor{
		outputId:    cfg.OutputId,
		nodeId:      cfg.NodeId,
		sinkId:      cfg.SinkId,
		leds:        cfg.Leds,
		average:     pixavg.New(windowSize, prev, 2),
		averageMagn: sliceavg.New(windowSize, make([]float64, cfg.Leds), 2),
		gradient:    grad,
		cfg:         cfg,
	}
}

func (proc *outputProcessor) process(freqs piecewiselinear.Function) []color.NRGBA {
	vals := make([]float64, 0, proc.leds*4)
	colors := make([]color.Color, 0, proc.leds)

	//for i := 0; i < seg.Leds/2+seg.Leds%2; i++ {
	//	magn := freqs.At(float64(i) / float64(seg.Leds-1))
	//
	//}

	//var lowFreqEnergy int

	mags := []float64{}
	for i := 0; i < proc.leds; i++ {
		magn := freqs.At(float64(i) / float64(proc.leds-1))
		mags = append(mags, magn)
	}

	proc.averageMagn.Add(mags)
	mags = proc.averageMagn.Current()

	for i := 0; i < proc.leds; i++ {
		//magn := freqs.At(float64(i) / float64(seg.Leds-1))
		magn := mags[i]

		c := proc.gradient.GetInterpolatedColor(magn)
		clr, _ := colorful.MakeColor(c)

		// Extract HSV color info, we'll use the Value to adjust the
		// brightness of the colors depending on frequency magnitude.
		hue, sat, val := clr.Hsv()

		// Easing effect easeOutCirc, ref: https://easings.net/#easeOutCirc
		// Should help exaggerate low values in magnitude e.g. high
		// frequency notes
		val = math.Sqrt(1 - math.Pow(magn-1, 2))
		//val = math.Sqrt(1 - math.Pow(magn-1, 3))
		//val = math.Sqrt(1 - math.Pow(val-1, 2))

		// adjust val partially based on peak magnitude
		peak := 1.0 // TODO @@@
		val = val * (0.25 + peak*2)
		val = math.Min(1, val) // prevent overflow

		//if i < 10 && val > 0.9 {
		//	lowFreqEnergy++
		//}
		// Adjust black point
		val = adjustBlackPoint(val, proc.blackPoint)

		// Convert the resulting color to RGBA
		hsv := colorful.Hsv(hue, sat, val)

		r, g, b, a := hsv.RGBA()

		vals = append(vals, float64(r), float64(g), float64(b), float64(a))
		colors = append(colors, hsv)
	}

	// Add the color data to the moving average accumulator for this segment
	proc.average.Add(colors)
	colors = proc.average.Current()

	if true {
		out := ""
		for _, c := range colors {
			r, g, b, _ := c.RGBA()
			out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprintf(" ")
		}
		fmt.Println(out)
	}

	//c1 := colors[0 : len(colors)/2]
	//c2 := make([]color.Color, len(c1))
	//copy(c2, c1)
	//reverse(c2)
	//colors = append(c1, c2...)

	colors2 := make([]color.NRGBA, len(colors))
	for i, c := range colors {
		r, g, b, a := c.RGBA()
		colors2[i] = color.NRGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: uint8(a >> 8),
		}
	}

	return colors2
}
