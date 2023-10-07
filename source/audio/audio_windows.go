package audio

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/cmplx"
	"time"

	"github.com/lucasb-eyer/go-colorful"

	"ledctl3/pkg/uuid"

	"ledctl3/pkg/audiocapture"
	"ledctl3/pkg/pixavg"
	"ledctl3/source/types"

	"github.com/VividCortex/ewma"
	"github.com/sgreben/piecewiselinear"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"

	"ledctl3/pkg/gradient"
)

const noiseCutoff = 1e-9

type Input struct {
	id uuid.UUID

	// ac is the audio capture device that captures desktop audio
	// and pipes samples captured during a specific time window
	ac *audiocapture.Capturer

	// events are passed to a consumer. do not overwrite, as the receiver won't
	// receive new events
	events chan types.UpdateEvent

	// outputs holds output-specific capture configurations
	outputs map[uuid.UUID]outputCaptureConfig
}

type outputCaptureConfig struct {
	id     uuid.UUID
	sinkId uuid.UUID
	leds   int
	colors gradient.Gradient

	// windowSize is the number of frames to average over
	windowSize int

	// blackPoint represents the normalization black point as a float value in
	// the range 0-1
	blackPoint float64

	// avg holds a moving array-based average for this output. The decay rate
	// is affected by windowSize.
	avg pixavg.MovingAverage

	// maxMagnAvg is a moving average of the maximum magnitude observed between
	// different audio frames. It helps make smoother transitions between
	// audio frames that have a frequently changing magnitude of the dominant
	// frequency. The decay rate is affected by windowSize.
	maxMagnAvg ewma.MovingAverage
}

func New() (*Input, error) {
	in := &Input{
		id:      uuid.New(),
		ac:      audiocapture.New(),
		events:  make(chan types.UpdateEvent),
		outputs: make(map[uuid.UUID]outputCaptureConfig),
	}

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

	return in, nil
}

func (in *Input) AssistedSetup() (map[string]any, error) {
	return map[string]any{
		"colors": []string{
			"#4a1524",
			"#065394",
			"#00b585",
			"#d600a4",
			"#ff004c",
		},
		"windowSize": 32,
		"blackPoint": 0.1,
	}, nil
}

func (in *Input) Id() uuid.UUID {
	return in.id
}

func (in *Input) Start(cfg types.InputConfig) error {
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

			// multiply windowSize by 8 to keep it more stable
			maxFreqAvg := ewma.NewMovingAverage(float64(windowSize) * 4)

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
				id:         out.Id,
				sinkId:     sinkCfg.Id,
				leds:       out.Leds,
				colors:     grad,
				blackPoint: blackPoint,
				avg:        avg,
				maxMagnAvg: maxFreqAvg,
			}
		}
	}

	err := in.ac.Start()
	if err != nil {
		return err
	}

	return nil
}

func (in *Input) Events() <-chan types.UpdateEvent {
	return in.events
}

func (in *Input) Stop() error {
	close(in.events)

	return in.ac.Stop()
}

// processFrame analyzes the audio frame, extracts frequency information and
// publishes an update event per-sink
func (in *Input) processFrame(samples []float64, peak float64) error {
	now := time.Now()

	if peak < noiseCutoff {
		outs := make(map[uuid.UUID][]types.UpdateEventOutput)

		for _, out := range in.outputs {
			c := out.colors.GetInterpolatedColor(0)
			clr, _ := colorful.MakeColor(c)
			hue, sat, _ := clr.Hsv()

			colors := make([]color.Color, out.leds)

			for i := 0; i < out.leds; i++ {
				val := adjustBlackPoint(0, out.blackPoint)

				colors[i] = colorful.Hsv(hue, sat, val)
			}

			out.avg.Add(colors)
			colors = out.avg.Current()

			outs[out.sinkId] = append(outs[out.sinkId], types.UpdateEventOutput{
				OutputId: out.id,
				Pix:      colors,
			})
		}

		for sinkId, outs := range outs {
			in.events <- types.UpdateEvent{
				SinkId:  sinkId,
				Outputs: outs,
				Latency: time.Since(now),
			}
		}

		return nil
	}

	// extract frequency information
	fft := fourier.NewFFT(len(samples))
	coeffs := fft.Coefficients(nil, window.Hamming(samples))

	// only keep the real part of the fft, and remove frequencies between
	// 20 and 24 khz
	coeffs = coeffs[:len(coeffs)/2*20/24]

	freqs := make([]float64, len(coeffs))
	var maxMagn float64

	// derive the frequency magnitude per bucket, and find maximum magnitude
	for i, coeff := range coeffs {
		val := cmplx.Abs(coeff)
		freqs[i] = val
		maxMagn = math.Max(maxMagn, val)
	}

	outs := make(map[uuid.UUID][]types.UpdateEventOutput)

	for _, out := range in.outputs {
		out.maxMagnAvg.Add(maxMagn)
		maxMagn = out.maxMagnAvg.Value()

		// normalize magnitudes to 0-1
		for i, freq := range freqs {
			freqs[i] = normalize(freq, 0, maxMagn)
			freqs[i] = math.Min(freqs[i], 1)
		}

		// perform piecewise linear interpolation between frequencies.
		f := piecewiselinear.Function{Y: freqs}
		f.X = scaleLog(0, 1, len(f.Y))

		colors := make([]color.Color, out.leds)
		for i := 0; i < out.leds; i++ {
			magn := f.At(float64(i) / float64(out.leds-1))

			c := out.colors.GetInterpolatedColor(magn)
			clr, _ := colorful.MakeColor(c)

			hue, sat, val := clr.Hsv()

			// should help exaggerate low magnitudes e.g. high frequency notes
			val = easeOutCirc(magn)

			// adjust val partially based on peak
			val = val * (1 + peak)
			val = math.Min(1, val)

			val = adjustBlackPoint(val, out.blackPoint)

			colors[i] = colorful.Hsv(hue, sat, val)
		}

		out.avg.Add(colors)
		colors = out.avg.Current()

		outs[out.sinkId] = append(outs[out.sinkId], types.UpdateEventOutput{
			OutputId: out.id,
			Pix:      colors,
		})
	}

	for sinkId, outs := range outs {
		in.events <- types.UpdateEvent{
			SinkId:  sinkId,
			Outputs: outs,
			Latency: time.Since(now),
		}
	}

	return nil
}

// easeOutCirc ref: https://easings.net/#easeOutCirc
func easeOutCirc(x float64) float64 {
	return math.Sqrt(1 - math.Pow(x-1, 2))
}

// adjustBlackPoint scales the (HSV) value of the color based on the passed
// minimum value
func adjustBlackPoint(v, min float64) float64 {
	return v*(1-min) + min
}

// normalize scales a value from min,max to 0,1
func normalize(val, min, max float64) float64 {
	if max == min {
		return max
	}

	return (val - min) / (max - min)
}

// scaleLog is used to scale frequencies logarithmically so that low ones are
// more pronounced
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
