package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"math/cmplx"
	"sync"
	"time"
	"unsafe"

	"github.com/VividCortex/ewma"
	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	gcolor "github.com/gookit/color"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/sgreben/piecewiselinear"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"

	"ledctl3/node/types"
	"ledctl3/pkg/gradient"
	"ledctl3/pkg/pixavg"
	"ledctl3/pkg/sliceavg"
)

type input struct {
	id     string
	name   string
	mmd    *wca.IMMDevice
	volume *wca.IAudioEndpointVolume
	level  float32

	cfg InputConfig

	captureWg     sync.WaitGroup
	captureCancel context.CancelFunc

	processWg     sync.WaitGroup
	processCancel context.CancelFunc

	procs  map[uuid.UUID]*outputProcessor
	frames chan []float64

	// freqMax is a moving average of the maximum magnitude observed between
	// different audio frames. It helps make smoother transitions between
	// audio frames that have a frequently changing magnitude of the dominant
	// frequency. The decay rate is affected by windowSize.
	freqMax ewma.MovingAverage
}

type outputProcessor struct {
	leds int

	//average holds a pixavg.Average for each segment. The decay rate
	//is affected by windowSize.
	average pixavg.MovingAverage

	averageMagn sliceavg.Average

	gradient gradient.Gradient

	// blackPoint represents the normalization black point as a float value in
	// the range 0-1
	blackPoint float64

	windowSize int // TODO
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
		leds:        cfg.Leds,
		average:     pixavg.New(windowSize, prev, 2),
		averageMagn: sliceavg.New(windowSize, make([]float64, cfg.Leds), 2),
		gradient:    grad,
	}
}

func (in *input) updateVolume() {
	var level float32
	if err := in.volume.GetMasterVolumeLevelScalar(&level); err != nil {
		fmt.Println("warn: Failed to get session volume", "error", err)
	}

	in.level = level
}

func (in *input) close() {
	fmt.Println("debug: Releasing audio session")

	in.mmd.Release()
	//in.volume.Release()
}

type InputConfig struct {
}

func (c InputConfig) Equal(c2 InputConfig) bool {
	return false
}

func newInput(id, name string, mmd *wca.IMMDevice) (*input, error) {
	var volume *wca.IAudioEndpointVolume
	if err := mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &volume); err != nil {
		fmt.Println("warn: Failed to activate AudioEndpointVolume for main session", "error", err)
		return nil, fmt.Errorf("activate main session: %w", err)
	}

	windowSize := 40

	in := &input{
		id:     id,
		name:   name,
		mmd:    mmd,
		volume: volume,
		procs:  make(map[uuid.UUID]*outputProcessor),
		frames: make(chan []float64),

		cfg: InputConfig{},

		freqMax: ewma.NewMovingAverage(float64(windowSize)),
	}

	//leds := 180
	//prev := make([]color.Color, leds)
	//for i := 0; i < len(prev); i++ {
	//	prev[i] = color.RGBA{}
	//}
	//
	//// TODO @@@ colors
	//grad, err := gradient.New(
	//	color.NRGBA{R: 0x4a, G: 0x15, B: 0x24, A: 0xff}, // "#4a1524",
	//	color.NRGBA{R: 0x06, G: 0x53, B: 0x94, A: 0xff}, // "#065394",
	//	color.NRGBA{R: 0x00, G: 0xb5, B: 0x85, A: 0xff}, // "#00b585",
	//	color.NRGBA{R: 0xd6, G: 0x60, B: 0xa4, A: 0xff}, // "#d600a4",
	//	color.NRGBA{R: 0xff, G: 0x00, B: 0x4c, A: 0xff}, // "#ff004c"
	//)
	//if err != nil {
	//	panic(err)
	//}
	//
	//in.procs[uuid.Nil] = &outputProcessor{
	//	leds:        leds,
	//	average:     pixavg.New(windowSize, prev, 2),
	//	averageMagn: sliceavg.New(windowSize, make([]float64, leds), 2),
	//	gradient:    grad,
	//}

	return in, nil
}

func (in *input) Configure(b []byte) error {
	var cfg InputConfig
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	if in.cfg.Equal(cfg) {
		return nil
	}

	in.stopCapture()
	//in.lastFrame = nil

	in.cfg = cfg

	in.startCapture()
	return nil
}

type OutputConfig struct {
	OutputId uuid.UUID
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   LinkConfig
}

type LinkConfig struct {
}

func (in *input) Connect(outsCfg []types.OutputConfig) error {
	var outs []OutputConfig

	for _, out := range outsCfg {
		var ocfg LinkConfig
		err := json.Unmarshal(out.Config, &ocfg)
		if err != nil {
			return err
		}

		outs = append(outs, OutputConfig{
			OutputId: out.OutputId,
			NodeId:   out.NodeId,
			SinkId:   out.SinkId,
			Leds:     out.Leds,
			Config:   ocfg,
		})
	}

	in.stopProcess()
	in.stopCapture()

	windowSize := 20

	for _, out := range outs {
		in.procs[out.OutputId] = newOutputProcessor(out, windowSize)
	}

	in.startProcess()
	in.startCapture()

	return nil
}

func (in *input) startCapture() {
	fmt.Println("startCapture", in.id)
	defer fmt.Println("startCapture done", in.id)

	ctx, cancel := context.WithCancel(context.Background())
	in.captureCancel = cancel

	in.captureWg.Add(1)
	go func() {
		defer in.captureWg.Done()

		err := in.capture(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}()
}

func (in *input) stopCapture() {
	fmt.Println("stopCapture", in.id)
	defer fmt.Println("stopCapture done", in.id)

	in.captureCancel()
	in.captureWg.Wait()
}

func (in *input) startProcess() {
	fmt.Println("startProcess", in.id)
	defer fmt.Println("startProcess done", in.id)

	if in.processCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	in.processCancel = cancel

	in.processWg.Add(1)
	go in.process(ctx)
}

func (in *input) stopProcess() {
	fmt.Println("stopProcess", in.id)
	defer fmt.Println("stopProcess done", in.id)

	if in.processCancel == nil {
		return
	}

	in.processCancel()
	in.processWg.Wait()

	in.procs = make(map[uuid.UUID]*outputProcessor)

	in.processCancel = nil
}

func (in *input) process(ctx context.Context) {
	defer in.processWg.Done()

	//if in.lastFrame != nil {
	//	in.processFrame(in.lastFrame)
	//}

	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-in.frames:
			//in.lastFrame = frame

			if !ok {
				panic("frames closed")
			}

			in.processFrame(frame)
		}
	}
}

func (in *input) processFrame(samples []float64) {
	// Extract frequency magnitudes using a fast fourier transform
	fft := fourier.NewFFT(len(samples))
	coeffs := fft.Coefficients(nil, window.Hamming(samples))

	// Only keep the real part of the fft, and also remove frequencies between
	// 19.2~ and 24 khz. x a 2 * 0.8 --> x * 2 / 5
	//coeffs = coeffs[:len(coeffs)*2/5]
	coeffs = coeffs[:len(coeffs)*4/6]

	// Get a logarithmic piecewise-interpolated projection of the frequencies
	freqs := in.calculateFrequencies(coeffs)

	for _, proc := range in.procs {
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
	}
}

func adjustBlackPoint(v, min float64) float64 {
	return v*(1-min) + min
}

func (in *input) calculateFrequencies(coeffs []complex128) piecewiselinear.Function {
	freqs := make([]float64, len(coeffs))
	var maxFreq float64

	// Keep the first part of the FFT. Also calculate the maximum magnitude
	// for this frame
	for i, coeff := range coeffs {
		val := cmplx.Abs(coeff)

		freqs[i] = val

		maxFreq = math.Max(maxFreq, val)
	}

	// Add an entry to the maxFrequency average accumulator
	in.freqMax.Add(maxFreq)
	maxFreq = in.freqMax.Value()

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
	x := make([]float64, nPoints)
	min, max = math.Min(max, min), math.Max(max, min)
	d := max - min
	for i := range x {
		v := min + d*(float64(i)/float64(nPoints-1))
		//v = math.Pow(v, 0.5)
		//v = math.Pow(v, 0.33)
		v = math.Pow(v, 0.4)
		x[i] = v
	}
	return x
}

func (in *input) run2(parentCtx context.Context) {
	fmt.Println("input.run", in.id)

	in.startProcess()
	in.startCapture()

	select {
	case <-parentCtx.Done():
		fmt.Println("parent context done", parentCtx.Err())
	}

	fmt.Println("PARENT DONE", in.id)

	//select {
	//case <-in.stop:
	//	fmt.Println("close with error")
	//	//case <-parentCtx.Done():
	//}

	in.stopCapture()
	in.stopProcess()

	in.close()
}

func (in *input) capture(ctx context.Context) error {
	var audioCli *wca.IAudioClient
	err := in.mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioCli)
	if err != nil {
		return err
	}

	var wfx *wca.WAVEFORMATEX
	err = audioCli.GetMixFormat(&wfx)
	if err != nil {
		return err
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	var defaultPeriod wca.REFERENCE_TIME
	var minimumPeriod wca.REFERENCE_TIME
	var latency time.Duration

	err = audioCli.GetDevicePeriod(&defaultPeriod, &minimumPeriod)
	if err != nil {
		return err
	}
	latency = time.Duration(int(defaultPeriod) * 100)

	wfx.NChannels = 2 // force stereo
	wfx.WFormatTag = 1
	wfx.WBitsPerSample = 32
	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
	wfx.CbSize = 0

	err = audioCli.Initialize(
		wca.AUDCLNT_SHAREMODE_SHARED,
		wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK,
		defaultPeriod, 0, wfx, nil,
	)
	if err != nil {
		return err
	}

	audioReadyEvent := wca.CreateEventExA(
		0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE,
	)
	defer wca.CloseHandle(audioReadyEvent)

	err = audioCli.SetEventHandle(audioReadyEvent)
	if err != nil {
		return err
	}

	var bufferFrameSize uint32
	err = audioCli.GetBufferSize(&bufferFrameSize)
	if err != nil {
		return err
	}

	var acc *wca.IAudioCaptureClient
	err = audioCli.GetService(wca.IID_IAudioCaptureClient, &acc)
	if err != nil {
		return err
	}
	defer acc.Release()

	_ = latency
	//fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	//fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	//fmt.Printf("Channels: %d\n", wfx.NChannels)
	//
	//fmt.Println("Default period: ", defaultPeriod)
	//fmt.Println("Minimum period: ", minimumPeriod)
	//fmt.Println("Latency: ", latency)
	//
	//fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	err = audioCli.Start()
	if err != nil {
		return fmt.Errorf("failed to start audio client: %w", err)
	}
	defer audioCli.Stop()

	var offset int
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64

	errorChan := make(chan error, 1)

	var isCapturing = true

	for {
		if !isCapturing {
			close(errorChan)
			break
		}
		go func() {
			errorChan <- watchEvent(ctx, audioReadyEvent)
		}()

		select {
		case <-ctx.Done():
			isCapturing = false
			<-errorChan
			return ctx.Err()
		case err := <-errorChan:
			if err != nil {
				isCapturing = false
				break
			}
			err = acc.GetBuffer(
				&data, &availableFrameSize, &flags,
				&devicePosition, &qcpPosition,
			)

			if err != nil {
				continue
			}

			if availableFrameSize == 0 {
				continue
			}

			start := unsafe.Pointer(data)
			if start == nil {
				continue
			}

			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
			buf := make([]byte, lim)

			for n := 0; n < lim; n++ {
				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
				buf[n] = *b
			}

			// close the buffer as soon as we extract the audio Samples
			err = acc.ReleaseBuffer(availableFrameSize)
			if err != nil {
				return fmt.Errorf("failed to release buffer: %w", err)
			}

			offset += lim

			samples := make([]float64, len(buf)/4)
			for i := 0; i < len(buf); i += 4 {
				v := float64(readInt32(buf[i : i+4]))
				samples = append(samples, v)
			}

			//fmt.Print(".")

			select {
			case in.frames <- samples:
			default:
			}

			// TODO: calculate impact of this call
			//var peak float32
			//err = ami.GetPeakValue(&peak)
			//if err != nil {
			//	continue
			//}
			//Peak = 1

			// Dispatch the received frame for processing. If the work queue
			// is full, this will block until c previous frame is processed.

			//fmt.Println(sess.id, "frame")
			//time.Sleep(500 * time.Millisecond)

		}
	}

	return nil
}

func watchEvent(ctx context.Context, event uintptr) (err error) {
	errorChan := make(chan error, 1)
	go func() {
		errorChan <- eventEmitter(event)
	}()
	select {
	case err = <-errorChan:
		close(errorChan)
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	}
}

func eventEmitter(event uintptr) (err error) {
	dw := wca.WaitForSingleObject(event, wca.INFINITE)
	if dw != 0 {
		return fmt.Errorf("failed to watch event")
	}
	return nil
}

// readInt32 reads a signed integer from a byte slice. only a slice with len(4)
// should be passed. equivalent to int32(binary.LittleEndian.Uint32(b))
func readInt32(b []byte) int32 {
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
}
