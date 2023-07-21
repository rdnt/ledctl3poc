package audio

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/cmplx"
	"time"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"

	"ledctl3/pkg/audiocapture"
	"ledctl3/pkg/pixavg"
	"ledctl3/source/types"

	"github.com/VividCortex/ewma"
	"github.com/pkg/errors"
	"github.com/sgreben/piecewiselinear"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"

	"ledctl3/pkg/gradient"
)

type Input struct {
	id uuid.UUID

	// ac is the audio capture device that captures desktop audio
	// and pipes samles captured during a specific time window
	ac *audiocapture.Capturer

	// events are passed to a consumer. do not overwrite, as the receiver won't
	// receive new events
	events chan types.UpdateEvent

	// maxLedCount holds the maximum number of LEDs across all segments.
	// It is updater every time we start a new audio capture session.
	maxLedCount int

	// stats holds timing and other useful info for the capture process
	stats Statistics

	// outputs holds output-specific capture configurations
	outputs map[uuid.UUID]outputCaptureConfig
}

type outputCaptureConfig struct {
	id     uuid.UUID
	sinkId uuid.UUID
	leds   int

	// colors is the color gradient to use for this output
	colors gradient.Gradient

	// windowSize is the number of frames to average over
	windowSize int

	// blackPoint represents the normalization black point as a float value in
	// the range 0-1
	blackPoint float64

	// avg holds a moving array-based average for this output. The decay rate
	// is affected by windowSize.
	avg pixavg.Average

	// maxFreqAvg is a moving average of the maximum magnitude observed between
	// different audio frames. It helps make smoother transitions between
	// audio frames that have a frequently changing magnitude of the dominant
	// frequency. The decay rate is affected by windowSize.
	maxFreqAvg ewma.MovingAverage
}

func New() (in *Input, err error) {
	in = new(Input)
	in.id = uuid.New()
	in.ac = audiocapture.New()

	go func() {
		for {
			select {
			case fr := <-in.ac.Frames():
				err := in.processFrame(fr.Samples, fr.Peak)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	in.events = make(chan types.UpdateEvent)

	//in.average = make(map[int]sliceewma.MovingAverage, len(in.segments))

	//in.maxFreqAvg = ewma.NewMovingAverage(float64(in.windowSize) * 8)

	//in.average = make(map[uuid.UUID]pixavg.Average, len(in.segments))
	//
	//for _, seg := range in.segments {
	//	prev := make([]color.Color, seg.Leds)
	//	for i := 0; i < len(prev); i++ {
	//		prev[i] = color.RGBA{}
	//	}
	//	in.average[seg.OutputId] = pixavg.New(in.windowSize, prev, 2)
	//}

	return in, nil
}

func (in *Input) AssistedSetup() (map[string]any, error) {
	return map[string]any{
		"colors": []string{
			//"#ffaeff",
			//"#9bbcff",
			//"#94fbd6",
			"#4a1524",
			"#065394",
			"#00b585",
			"#d600a4",
			"#ff004c",
		},
		"windowSize": 20,
		"blackPoint": 0.2,
	}, nil
}

func (in *Input) Id() uuid.UUID {
	return in.id
}

type Statistics struct {
	BitRate int // in hz
	Latency time.Duration
}

func (in *Input) Statistics() Statistics {
	return Statistics{}
}

func (in *Input) Start(cfg types.SinkConfig) error {
	fmt.Printf("## starting audio source with config: %#v\n", cfg)

	in.outputs = make(map[uuid.UUID]outputCaptureConfig)

	for _, sinkCfg := range cfg.Sinks {
		for _, out := range sinkCfg.Outputs {
			windowSize := out.Config["windowSize"].(int)
			blackPoint := out.Config["blackPoint"].(float64)

			var colors []color.Color
			for _, hex := range out.Config["colors"].([]string) {
				clr, err := colorful.Hex(hex)
				if err != nil {
					return err
				}

				colors = append(colors, clr)
			}

			maxFreqAvg := ewma.NewMovingAverage(float64(windowSize))

			prev := make([]color.Color, out.Leds)
			for i := 0; i < len(prev); i++ {
				prev[i] = color.RGBA{}
			}
			avg := pixavg.New(windowSize, prev, 2)

			grad, err := gradient.New(colors...)
			if err != nil {
				return err
			}

			in.outputs[out.Id] = outputCaptureConfig{
				id:     out.Id,
				sinkId: sinkCfg.Id,
				leds:   out.Leds,
				colors: grad,
				//windowSize: windowSize,
				blackPoint: blackPoint,
				avg:        avg,
				maxFreqAvg: maxFreqAvg,
			}
		}
	}

	err := in.ac.Start()
	if errors.Is(err, context.Canceled) {
		return err
	} else if err != nil {
		log.Println(err)
		time.Sleep(1 * time.Second)
	}

	return nil
}

func (in *Input) Events() <-chan types.UpdateEvent {
	return in.events
}

func (in *Input) Stop() error {
	//if in.cancel == nil {
	//	return nil
	//}
	//
	//in.cancel()
	//in.cancel = nil
	//
	//<-in.done

	return nil
}

// processFrame analyses the audio frame, extracts frequency information and
// creates the necessary update event
func (in *Input) processFrame(samples []float64, peak float64) error {
	now := time.Now()

	if peak < 1e-9 {
		// skip calculations, set all frequencies to 0

		segs := make(map[uuid.UUID][]types.UpdateEventOutput)

		for _, out := range in.outputs {
			colors := make([]color.Color, out.leds)
			for i := 0; i < out.leds; i++ {
				clr, _ := colorful.MakeColor(out.colors.GetInterpolatedColor(0))
				hue, sat, _ := clr.Hsv()
				val := adjustBlackPoint(0, out.blackPoint)
				hsv := colorful.Hsv(hue, sat, val)

				colors[i] = hsv
			}

			out.avg.Add(colors)
			colors = out.avg.Current()

			segs[out.sinkId] = append(segs[out.sinkId], types.UpdateEventOutput{
				OutputId: out.id,
				Pix:      colors,
			})
		}

		for sinkId, outs := range segs {
			in.events <- types.UpdateEvent{
				SinkId:  sinkId,
				Outputs: outs,
				Latency: time.Since(now),
			}
		}

		return nil
	}

	// Extract frequency magnitudes using in fast fourier transform
	fft := fourier.NewFFT(len(samples))
	coeffs := fft.Coefficients(nil, window.Hamming(samples))

	// Only keep the real part of the fft, and also remove frequencies between
	// 19.2~ and 24 khz. x / 2 * 0.8 --> x * 2 / 5
	coeffs = coeffs[:len(coeffs)*2/5]

	// Get in logarithmic piecewise-interpolated projection of the frequencies
	freqs := in.calculateFrequencies(coeffs)

	segs := make(map[uuid.UUID][]types.UpdateEventOutput)

	for _, out := range in.outputs {
		vals := make([]float64, 0, out.leds*4)
		colors := make([]color.Color, 0, out.leds)

		for i := 0; i < out.leds; i++ {
			magn := freqs.At(float64(i) / float64(out.leds-1))

			c := out.colors.GetInterpolatedColor(magn)
			clr, _ := colorful.MakeColor(c)

			// Extract HSV color info, we'll use the Value to adjust the
			// brightness of the colors depending on frequency magnitude.
			hue, sat, val := clr.Hsv()

			// Easing effect easeOutCirc, ref: https://easings.net/#easeOutCirc
			// Should help exaggerate low values in magnitude e.g. high
			// frequency notes
			val = math.Sqrt(1 - math.Pow(magn-1, 2))

			// adjust val partially based on peak magnitude
			val = val * (1 + peak)
			val = math.Min(1, val) // prevent overflow

			// Adjust black point
			val = adjustBlackPoint(val, out.blackPoint)

			// Convert the resulting color to RGBA
			hsv := colorful.Hsv(hue, sat, val)

			r, g, b, a := hsv.RGBA()

			vals = append(vals, float64(r), float64(g), float64(b), float64(a))
			colors = append(colors, hsv)
		}

		// Add the color data to the moving average accumulator for this segment
		out.avg.Add(colors)
		colors = out.avg.Current()

		// Create the pix slice from the color data
		//pix := make([]uint8, len(colors))
		//for j := 0; j < len(vals); j++ {
		//	pix[j] = uint8(uint16(vals[j]) >> 8)
		//}

		//pix := make([]color.Color, len(colors)*4)
		//
		//for _, c := range colors {
		//	r, g, b, in := c.RGBA()
		//	pix = append(pix,
		//		uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(in>>8),
		//	)
		//}

		// DEBUG
		//if seg.OutputId == 0 {
		// TODO: COMMENT OUT
		//out := ""
		//for _, c := range colors {
		//	r, g, b, _ := c.RGBA()
		//	out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprintf(" ")
		//}
		//fmt.Println(out)
		////}

		segs[out.sinkId] = append(segs[out.sinkId], types.UpdateEventOutput{
			OutputId: out.id,
			Pix:      colors,
		})
	}

	for sinkId, outs := range segs {
		in.events <- types.UpdateEvent{
			SinkId:  sinkId,
			Outputs: outs,
			Latency: time.Since(now),
		}
	}

	return nil
}

func adjustBlackPoint(v, min float64) float64 {
	return v*(1-min) + min
}

func (in *Input) calculateFrequencies(coeffs []complex128) piecewiselinear.Function {
	freqs := make([]float64, len(coeffs))
	var maxFreq float64

	// Keep the first part of the FFT. Also calculate the maximum magnitude
	// for this frame
	for i, coeff := range coeffs {
		val := cmplx.Abs(coeff)

		freqs[i] = val

		maxFreq = math.Max(maxFreq, val)
	}

	// TODO: per-output
	//// Add an entry to the maxFrequency average accumulator
	//in.freqMax.Add(maxFreq)
	//maxFreq = in.freqMax.Value()

	// Normalize frequencies between [0,1] based on maxFreq
	for i, freq := range freqs {
		freqs[i] = normalize(freq, 0, maxFreq)
		freqs[i] = math.Min(freqs[i], 1)
	}

	// Perform piecewise linear interpolation between frequencies. Also scale
	// frequencies logarithmically so that low ones are more pronounced.
	f := piecewiselinear.Function{Y: freqs}
	f.X = scaleLog(0, 1, len(f.Y))

	return f
}

// normalize scales a value from min,max to 0,1
func normalize(val, min, max float64) float64 {
	if max == min {
		return max
	}

	return (val - min) / (max - min)
}

func scaleLog(min, max float64, nPoints int) []float64 {
	X := make([]float64, nPoints)
	min, max = math.Min(max, min), math.Max(max, min)
	d := max - min
	for i := range X {
		v := min + d*(float64(i)/float64(nPoints-1))
		v = math.Pow(v, 0.5)
		X[i] = v
	}
	return X
}
