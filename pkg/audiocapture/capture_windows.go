package audiocapture

import (
	"context"
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
	"github.com/pkg/errors"

	wcaami "ledctl3/_source-old/audio/wca-ami"
)

// Capturer is an audio capture device. It is NOT safe for concurrent use.
// Use the Frames() method to receive a channel with captured audio frames.
type Capturer struct {
	cancel        context.CancelFunc
	cancelCapture context.CancelFunc
	done          chan bool
	frames        chan Frame
}

// Frame represents an audio frame
type Frame struct {
	// Samples is a collection of PCM samples encoded as float64
	Samples []float64
	// Peak is the peak audio meter value for this frame (0-1)
	Peak float64
}

func New() *Capturer {
	return &Capturer{
		frames: make(chan Frame),
	}
}

func (c *Capturer) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.done = make(chan bool)

	go func() {
		// TODO: return error early if capture doesnt even start (e.g. move outside of goroutine)
		err := c.startCapture(ctx)
		if errors.Is(err, context.Canceled) {
			c.done <- true
			return
		} else if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}

func (c *Capturer) Stop() error {
	c.cancel()
	<-c.done

	return nil
}

func (c *Capturer) Restart() error {
	c.cancel()
	<-c.done

	return c.Start()
}

func (c *Capturer) Frames() <-chan Frame {
	return c.frames
}

func (c *Capturer) startCapture(ctx context.Context) error {
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

	var audioCli *wca.IAudioClient
	err = mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioCli)
	if err != nil {
		return err
	}
	defer audioCli.Release()

	var ami *wcaami.IAudioMeterInformation
	err = mmd.Activate(
		wca.IID_IAudioMeterInformation, wca.CLSCTX_ALL, nil, &ami,
	)
	if err != nil {
		return err
	}
	defer ami.Release()

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
	defer func() {
		_ = wca.CloseHandle(audioReadyEvent)
	}()

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

	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
	fmt.Printf("Channels: %d\n", wfx.NChannels)

	fmt.Println("Default period: ", defaultPeriod)
	fmt.Println("Minimum period: ", minimumPeriod)
	fmt.Println("Latency: ", latency)

	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)

	err = audioCli.Start()
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

			// Release the buffer as soon as we extract the audio Samples
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
			//Peak = 1

			// Dispatch the received frame for processing. If the work queue
			// is full, this will block until c previous frame is processed.
			c.frames <- Frame{
				Samples: samples,
				Peak:    float64(peak),
			}
		}
	}

	err = audioCli.Stop()
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
