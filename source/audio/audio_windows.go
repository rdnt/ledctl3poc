package audio

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/cmplx"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"

	"ledctl3/pkg/pixavg"
	"ledctl3/source/types"

	wca_ami "ledctl3/source/audio/wca-ami"

	"github.com/VividCortex/ewma"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/pkg/errors"
	"github.com/sgreben/piecewiselinear"
	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/gonum/dsp/window"

	"ledctl3/pkg/gradient"
)

type Capture struct {
	// events are passed to a consumer. do not overwrite, as the receiver won't
	// receive new events
	events chan types.UpdateEvent

	shutdown      context.CancelFunc
	cancelCapture context.CancelFunc
	done          chan bool

	// maxLedCount holds the maximum number of LEDs across all segments.
	maxLedCount int

	frames chan frame

	gradient   gradient.Gradient
	windowSize int

	stats Statistics

	// average holds a pixavg.Average for each segment. The decay rate
	// is affected by windowSize.
	average map[uuid.UUID]pixavg.Average

	// freqMax is a moving average of the maximum magnitude observed between
	// different audio frames. It helps make smoother transitions between
	// audio frames that have a frequently changing magnitude of the dominant
	// frequency. The decay rate is affected by windowSize.
	freqMax ewma.MovingAverage

	// blackPoint represents the normalization black point as a float value in
	// the range 0-1
	blackPoint float64

	sinkCfg types.SinkConfig
	id      uuid.UUID
}

func (a *Capture) AssistedSetup() (map[string]any, error) {
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
		"windowSize": 40,
		"blackPoint": 0.2,
	}, nil
}

func (a *Capture) Id() uuid.UUID {
	return a.id
}

type Segment struct {
	SinkId   uuid.UUID
	OutputId uuid.UUID
	Leds     int
}

// frame represents an audio frame
type frame struct {
	// samples is a collection of PCM samples encoded as float64
	samples []float64
	// peak is the peak audio meter value for this frame
	peak float64
}

type Statistics struct {
	BitRate int // in hz
	Latency time.Duration
}

func (a *Capture) Statistics() Statistics {
	return Statistics{}
}

func (a *Capture) Start(cfg types.SinkConfig) error {
	a.sinkCfg = cfg

	ctx, cancel := context.WithCancel(context.Background())
	a.shutdown = cancel
	a.done = make(chan bool)

	fmt.Printf("## starting audio source with config: %#v\n", cfg)

	//segs := make([]Segment, 0)
	//for _, sinkCfg := range cfg.Sinks {
	//	for _, output := range sinkCfg.Outputs {
	//		segs = append(segs, Segment{
	//			SinkId:   sinkCfg.Id,
	//			OutputId: output.Id,
	//			Leds:     output.Leds,
	//		})
	//	}
	//}

	a.average = make(map[uuid.UUID]pixavg.Average, len(segs))

	for _, seg := range segs {
		prev := make([]color.Color, seg.Leds)
		for i := 0; i < len(prev); i++ {
			prev[i] = color.RGBA{}
		}
		a.average[seg.OutputId] = pixavg.New(a.windowSize, prev, 2)
	}

	//_ = WithSegments(segs)(a)
	a.frames = make(chan frame)

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Print("========================= startCapture stopped\n")
				a.done <- true
				return
			default:
				var childCtx context.Context
				childCtx, a.cancelCapture = context.WithCancel(ctx)

				err := a.startCapture(childCtx)
				if errors.Is(err, context.Canceled) {
					return
				} else if err != nil {
					log.Println(err)
					time.Sleep(1 * time.Second)
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case frame := <-a.frames:
				err := a.processFrame(frame.samples, frame.peak)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	return nil
}

func (a *Capture) Events() chan types.UpdateEvent {
	return a.events
}

func (a *Capture) Stop() error {
	if a.shutdown == nil {
		return nil
	}

	a.shutdown()
	a.shutdown = nil

	<-a.done

	return nil
}

func (a *Capture) startCapture(ctx context.Context) error {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	err = wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator, &mmde,
	)
	if err != nil {
		return err
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	err = mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd)
	if err != nil {
		return err
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	err = mmd.OpenPropertyStore(wca.STGM_READ, &ps)
	if err != nil {
		return err
	}
	defer ps.Release()

	var ac *wca.IAudioClient
	err = mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &ac)
	if err != nil {
		return err
	}
	defer ac.Release()

	var ami *wca_ami.IAudioMeterInformation
	err = mmd.Activate(
		wca.IID_IAudioMeterInformation, wca.CLSCTX_ALL, nil, &ami,
	)
	if err != nil {
		return err
	}
	defer ami.Release()

	var wfx *wca.WAVEFORMATEX
	err = ac.GetMixFormat(&wfx)
	if err != nil {
		return err
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	var defaultPeriod wca.REFERENCE_TIME
	var minimumPeriod wca.REFERENCE_TIME
	var latency time.Duration

	err = ac.GetDevicePeriod(&defaultPeriod, &minimumPeriod)
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

	a.stats.Latency = latency
	a.stats.BitRate = int(wfx.NSamplesPerSec)

	err = ac.Initialize(
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

	err = ac.SetEventHandle(audioReadyEvent)
	if err != nil {
		return err
	}

	var bufferFrameSize uint32
	err = ac.GetBufferSize(&bufferFrameSize)
	if err != nil {
		return err
	}

	var acc *wca.IAudioCaptureClient
	err = ac.GetService(wca.IID_IAudioCaptureClient, &acc)
	if err != nil {
		return err
	}
	defer acc.Release()

	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	fmt.Printf("Channels: %d\n", wfx.NChannels)

	fmt.Println("Default period: ", defaultPeriod)
	fmt.Println("Minimum period: ", minimumPeriod)
	fmt.Println("Latency: ", latency)

	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	err = ac.Start()
	if err != nil {
		return err
	}

	var offset int
	var b *byte
	var data *byte
	var availableFrameSize uint32
	var flags uint32
	var devicePosition uint64
	var qcpPosition uint64

	errorChan := make(chan error, 1)

	var isCapturing = true

loop:
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
			break loop
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

			// Release the buffer as soon as we extract the audio samples
			err = acc.ReleaseBuffer(availableFrameSize)
			if err != nil {
				return errors.WithMessage(err, "failed to release buffer")
			}

			offset += lim

			samples := make([]float64, len(buf)/4)
			for i := 0; i < len(buf); i += 4 {
				v := float64(readInt32(buf[i : i+4]))
				samples = append(samples, v)
			}

			// TODO: calculate impact of this call
			var peak float32
			err = ami.GetPeakValue(&peak)
			if err != nil {
				continue
			}
			//peak = 1

			// Dispatch the received frame for processing. If the work queue
			// is full, this will block until a previous frame is processed.
			a.frames <- frame{
				samples: samples,
				peak:    float64(peak),
			}
		}
	}

	err = ac.Stop()
	if err != nil {
		return errors.Wrap(err, "failed to stop audio client")
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

// processFrame analyses the audio frame, extracts frequency information and
// creates the necessary update event
func (a *Capture) processFrame(samples []float64, peak float64) error {
	now := time.Now()

	if peak < 1e-9 {
		// skip calculations, set all frequencies to 0

		segs := make([]types.UpdateEventOutput, 0, len(a.segments))

		for _, seg := range a.segments {
			colors := make([]color.Color, seg.Leds)
			for i := 0; i < seg.Leds; i++ {
				colors[i] = color.RGBA{}
			}

			a.average[seg.OutputId].Add(colors)
			colors = a.average[seg.OutputId].Current()

			//if seg.OutputId ==  {
			//	out := ""
			//	for _, c := range colors {
			//		r, g, b, _ := c.RGBA()
			//		out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprintf(" ")
			//	}
			//	fmt.Println(out)
			//}

			segs = append(segs, types.UpdateEventOutput{
				OutputId: seg.OutputId,
				Pix:      colors,
			})
		}

		a.events <- types.UpdateEvent{
			SinkId:  a.segments[0].SinkId,
			Outputs: segs,
			Latency: time.Since(now),
		}

		return nil
	}

	// Extract frequency magnitudes using a fast fourier transform
	fft := fourier.NewFFT(len(samples))
	coeffs := fft.Coefficients(nil, window.Hamming(samples))

	// Only keep the real part of the fft, and also remove frequencies between
	// 19.2~ and 24 khz. x / 2 * 0.8 --> x * 2 / 5
	coeffs = coeffs[:len(coeffs)*2/5]

	// Get a logarithmic piecewise-interpolated projection of the frequencies
	freqs := a.calculateFrequencies(coeffs)

	segs := make([]types.UpdateEventOutput, 0, len(a.segments))

	for _, seg := range a.segments {
		vals := make([]float64, 0, seg.Leds*4)
		colors := make([]color.Color, 0, seg.Leds)

		for i := 0; i < seg.Leds; i++ {
			magn := freqs.At(float64(i) / float64(seg.Leds-1))

			c := a.gradient.GetInterpolatedColor(magn)
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
			val = adjustBlackPoint(val, a.blackPoint)

			// Convert the resulting color to RGBA
			hsv := colorful.Hsv(hue, sat, val)

			r, g, b, a := hsv.RGBA()

			vals = append(vals, float64(r), float64(g), float64(b), float64(a))
			colors = append(colors, hsv)
		}

		// Add the color data to the moving average accumulator for this segment
		a.average[seg.OutputId].Add(colors)
		colors = a.average[seg.OutputId].Current()

		// Create the pix slice from the color data
		//pix := make([]uint8, len(colors))
		//for j := 0; j < len(vals); j++ {
		//	pix[j] = uint8(uint16(vals[j]) >> 8)
		//}

		//pix := make([]color.Color, len(colors)*4)
		//
		//for _, c := range colors {
		//	r, g, b, a := c.RGBA()
		//	pix = append(pix,
		//		uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8),
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

		segs = append(segs, types.UpdateEventOutput{
			OutputId: seg.OutputId,
			Pix:      colors,
		})
	}

	a.events <- types.UpdateEvent{
		SinkId:  a.segments[0].SinkId,
		Outputs: segs,
		Latency: time.Since(now),
	}

	return nil
}

func adjustBlackPoint(v, min float64) float64 {
	return v*(1-min) + min
}

func (a *Capture) calculateFrequencies(coeffs []complex128) piecewiselinear.Function {
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
	a.freqMax.Add(maxFreq)
	maxFreq = a.freqMax.Value()

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

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func New(opts ...Option) (a *Capture, err error) {
	a = new(Capture)
	a.id = uuid.New()

	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return nil, err
		}
	}

	a.gradient, err = gradient.New(a.colors...)
	if err != nil {
		return nil, err
	}

	a.events = make(chan types.UpdateEvent, len(a.segments))

	//a.average = make(map[int]sliceewma.MovingAverage, len(a.segments))

	a.freqMax = ewma.NewMovingAverage(float64(a.windowSize) * 8)

	a.average = make(map[uuid.UUID]pixavg.Average, len(a.segments))

	for _, seg := range a.segments {
		prev := make([]color.Color, seg.Leds)
		for i := 0; i < len(prev); i++ {
			prev[i] = color.RGBA{}
		}
		a.average[seg.OutputId] = pixavg.New(a.windowSize, prev, 2)
	}

	return a, nil
}
