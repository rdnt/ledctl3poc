package audio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	"github.com/moutend/go-wca/pkg/wca"

	"ledctl3/node"
	"ledctl3/node/common"
)

func init() {
	node.RegisterSource("audio", New())
}

func New() node.Source {
	c := &source{
		inputs:       make(map[uuid.UUID]*input),
		updateVolume: make(chan bool, 1),
	}

	return c
}

type source struct {
	id    uuid.UUID
	reg   common.IORegistry
	state common.StateHolder

	// needed for device change notifications
	mmDeviceEnumerator   *wca.IMMDeviceEnumerator
	mmNotificationClient *wca.IMMNotificationClient

	// our main input and output inputs
	main *input

	inputs map[uuid.UUID]*input

	wg     sync.WaitGroup
	cancel context.CancelFunc

	updateVolume chan bool
}

func (s *source) Configure(cfg []byte) error {
	return nil
}

func (s *source) Close() error {
	s.stop()
	return nil
}

func (s *source) Id() uuid.UUID {
	return s.id
}

func (s *source) Init(id uuid.UUID, reg common.IORegistry, state common.StateHolder) error {
	s.id = id
	s.reg = reg
	s.state = state

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(1)
	go s.mainLoop(ctx)

	return nil
}

func (s *source) stop() {
	fmt.Println("capturer.stop")

	if s.cancel == nil {
		return
	}

	s.cancel()
	s.wg.Wait()

	ole.CoUninitialize()

	s.cancel = nil
}

func (s *source) init() error {
	// we must call this every time we're about to list devices, i think. could be wrong
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {

		// if the error is "Incorrect function" that corresponds to 0x00000001,
		// which represents E_FALSE in COM error handling. this is fine for this function,
		// and just means that the call was redundant.
		const eFalse = 1
		oleError := &ole.OleError{}

		if errors.As(err, &oleError) {
			if oleError.Code() == eFalse {
				fmt.Println("warn: CoInitializeEx failed with E_FALSE due to redundant invocation")
			} else {
				fmt.Println("warn: Failed to call CoInitializeEx",
					"isOleError", true,
					"error", err,
					"oleError", oleError)

				return fmt.Errorf("call CoInitializeEx: %w", err)
			}
		} else {
			fmt.Println("warn: Failed to call CoInitializeEx",
				"isOleError", false,
				"error", err,
				"oleError", nil)

			return fmt.Errorf("call CoInitializeEx: %w", err)
		}

	}

	// ensure we have a device enumerator
	if err := s.getDeviceEnumerator(); err != nil {
		fmt.Println("warn: Failed to get device enumerator", "error", err)
		return fmt.Errorf("get device enumerator: %w", err)
	}

	// receive notifications whenever the default device changes (only do this once)
	if s.mmNotificationClient == nil {
		if err := s.registerDefaultDeviceChangeCallback(); err != nil {
			fmt.Println("warn: Failed to register default device change callback", "error", err)
			return fmt.Errorf("register default device change callback: %w", err)
		}
	}

	return nil
}

func (s *source) start() {
	fmt.Println("capturer.start")

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(1)
	go s.mainLoop(ctx)

	return
}

func (s *source) restart() {
	fmt.Println("capturer.restart")
	s.stop()
	s.start()
}

func (s *source) mainLoop(ctx context.Context) {
	fmt.Println("capturer.mainLoop")

	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("capturer.mainLoop.iter")
			err := s.run(ctx)
			if err != nil {
				fmt.Println(err)
			}

			time.Sleep(30 * time.Millisecond)
		}
	}
}

func (s *source) run(ctx context.Context) error {
	fmt.Println("capturer.run")

	for id, sess := range s.inputs {
		sess.close()
		s.reg.RemoveInput(id)
		delete(s.inputs, id)
	}

	err := s.init()
	if err != nil {
		return err
	}

	err = s.initMainSession()
	if err != nil {
		return err
	}

	if err := s.initSessions(); err != nil {
		return err
	}

	captureCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var captureWg sync.WaitGroup

	for _, in := range s.inputs {
		captureWg.Add(1)

		go func() {
			defer captureWg.Done()
			defer cancel()

			in.run2(captureCtx)
		}()
	}

	captureWg.Wait()

	return ctx.Err()
}

func (s *source) initMainSession() error {
	// get the currently active default output device.
	var defaultOutputEndpoint *wca.IMMDevice
	if err := s.mmDeviceEnumerator.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &defaultOutputEndpoint); err != nil {
		fmt.Println("warn: Failed to call GetDefaultAudioEndpoint (out)", "error", err)
		return fmt.Errorf("call GetDefaultAudioEndpoint (out): %w", err)
	}
	defer defaultOutputEndpoint.Release()

	// create the main session
	master, err := newInput(uuid.Nil, s.id, "main", defaultOutputEndpoint)
	if err != nil {
		fmt.Println("warn: Failed to get main audio output session", "error", err)
		return fmt.Errorf("get main audio output session: %w", err)
	}

	s.main = master

	return nil
}

func (s *source) Release() error {
	// skip unregistering the mmnotificationclient, as it's not implemented in go-wca
	if s.mmDeviceEnumerator != nil {
		s.mmDeviceEnumerator.Release()
	}

	fmt.Println("debug: Released WCA session finder instance")

	return nil
}

func (s *source) getDeviceEnumerator() error {

	// get the IMMDeviceEnumerator (only once)
	if s.mmDeviceEnumerator == nil {
		if err := wca.CoCreateInstance(
			wca.CLSID_MMDeviceEnumerator,
			0,
			wca.CLSCTX_ALL,
			wca.IID_IMMDeviceEnumerator,
			&s.mmDeviceEnumerator,
		); err != nil {
			fmt.Println("warn: Failed to call CoCreateInstance", "error", err)
			return fmt.Errorf("call CoCreateInstance: %w", err)
		}
	}

	return nil
}

func (s *source) registerDefaultDeviceChangeCallback() error {
	s.mmNotificationClient = &wca.IMMNotificationClient{}

	s.mmNotificationClient = wca.NewIMMNotificationClient(wca.IMMNotificationClientCallback{
		OnDefaultDeviceChanged: func(flow wca.EDataFlow, role wca.ERole, pwstrDeviceId string) error {
			fmt.Println("default endpoint updated", flow, role, pwstrDeviceId)
			s.restart()

			//if s.captureCancel != nil {
			//	fmt.Println("capture canceled OnDefaultDeviceChanged")
			//	s.captureCancel()
			//}
			return nil
		},
		OnDeviceAdded: func(pwstrDeviceId string) error {
			fmt.Println("device added", pwstrDeviceId)
			s.restart()
			//if s.captureCancel != nil {
			//	fmt.Println("capture canceled OnDeviceAdded")
			//	s.captureCancel()
			//}
			return nil
		},
		OnDeviceRemoved: func(pwstrDeviceId string) error {
			fmt.Println("device remove", pwstrDeviceId)

			s.restart()
			//if s.captureCancel != nil {
			//	fmt.Println("capture canceled OnDeviceRemoved")
			//	s.captureCancel()
			//}
			return nil
		},
		OnDeviceStateChanged: func(pwstrDeviceId string, dwNewState uint64) error {
			fmt.Println("device state changed", pwstrDeviceId, dwNewState)

			s.restart()
			//if s.captureCancel != nil {
			//	fmt.Println("capture canceled OnDeviceStateChanged")
			//	s.captureCancel()
			//}
			return nil
		},
		OnPropertyValueChanged: func(pwstrDeviceId string, dwNewValue uint64) error {
			//fmt.Println("property value changed", pwstrDeviceId, dwNewValue)
			//s.updateVolume <- true
			return nil
		},
	})

	if err := s.mmDeviceEnumerator.RegisterEndpointNotificationCallback(s.mmNotificationClient); err != nil {
		return fmt.Errorf("call RegisterEndpointNotificationCallback: %w", err)
	}

	return nil
}

func (s *source) initSessions() error {
	// get list of devices
	var deviceCollection *wca.IMMDeviceCollection

	if err := s.mmDeviceEnumerator.EnumAudioEndpoints(wca.ERender, wca.DEVICE_STATE_ACTIVE, &deviceCollection); err != nil {
		fmt.Println("warn: Failed to enumerate active audio endpoints", "error", err)
		return fmt.Errorf("enumerate active audio endpoints: %w", err)
	}

	// check how many devices there are
	var deviceCount uint32

	if err := deviceCollection.GetCount(&deviceCount); err != nil {
		fmt.Println("warn: Failed to get device count from device collection", "error", err)
		return fmt.Errorf("get device count from device collection: %w", err)
	}

	// for each device:
	for deviceIdx := uint32(0); deviceIdx < deviceCount; deviceIdx++ {
		// get its IMMDevice instance
		var endpoint *wca.IMMDevice
		if err := deviceCollection.Item(deviceIdx, &endpoint); err != nil {
			fmt.Println("warn: Failed to get device from device collection",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d from device collection: %w", deviceIdx, err)
		}
		defer endpoint.Release()

		// get the device's property store
		var propertyStore *wca.IPropertyStore

		if err := endpoint.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
			fmt.Println("warn: Failed to open property store for endpoint",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("open endpoint %d property store: %w", deviceIdx, err)
		}
		defer propertyStore.Release()

		// query the property store for the device's description and friendly name
		value := &wca.PROPVARIANT{}

		if err := propertyStore.GetValue(&wca.PKEY_AudioEndpoint_GUID, value); err != nil {
			fmt.Println("warn: Failed to get guid for device",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d guid: %w", deviceIdx, err)
		}

		// device description i.e. "Headphones"
		endpointUuid, err := uuid.Parse(value.String())
		if err != nil {
			return fmt.Errorf("get device %d invalid uuid: %w", deviceIdx, err)
		}

		if err := propertyStore.GetValue(&wca.PKEY_Device_DeviceDesc, value); err != nil {
			fmt.Println("warn: Failed to get description for device",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d description: %w", deviceIdx, err)
		}

		// device description i.e. "Headphones"
		endpointDescription := strings.ToLower(value.String())

		if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value); err != nil {
			fmt.Println("warn: Failed to get friendly name for device",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d friendly name: %w", deviceIdx, err)
		}

		// device friendly name i.e. "Headphones (Realtek Audio)"
		endpointFriendlyName := value.String()

		fmt.Println("debug: Enumerated device info",
			"deviceIdx", deviceIdx,
			"deviceUuid", endpointUuid,
			"deviceDescription", endpointDescription,
			"deviceFriendlyName", endpointFriendlyName)

		in, err := newInput(endpointUuid, s.id, endpointFriendlyName, endpoint)
		if err != nil {
			fmt.Println("warn: Failed to get main session for device",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d main session: %w", deviceIdx, err)
		}

		s.reg.AddInput(in)

		s.inputs[endpointUuid] = in
	}

	return nil
}

//func (s *source) Configure(b []byte) error {
//	return nil
//}
//
//func (s *source) Close() error {
//	return nil
//}
//
//type frame struct {
//	// Samples is a collection of PCM samples encoded as float64
//	Samples []float64
//	// Peak is the peak audio meter value for this frame (0-1)
//	Peak float64
//}
//
//func (sf *source) init() ([]any, error) {
//	//sessions := []any{}
//
//	// we must call this every time we're about to list devices, i think. could be wrong
//	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
//
//		// if the error is "Incorrect function" that corresponds to 0x00000001,
//		// which represents E_FALSE in COM error handling. this is fine for this function,
//		// and just means that the call was redundant.
//		const eFalse = 1
//		oleError := &ole.OleError{}
//
//		if errors.As(err, &oleError) {
//			if oleError.Code() == eFalse {
//				fmt.Println("warn: CoInitializeEx failed with E_FALSE due to redundant invocation")
//			} else {
//				return nil, fmt.Errorf("call CoInitializeEx: %w", err)
//			}
//		} else {
//			return nil, fmt.Errorf("call CoInitializeEx: %w", err)
//		}
//
//	}
//	defer ole.CoUninitialize()
//
//	// ensure we have a device enumerator
//	if err := sf.getDeviceEnumerator(); err != nil {
//		return nil, fmt.Errorf("get device enumerator: %w", err)
//	}
//
//	// receive notifications whenever the default device changes (only do this once)
//	if sf.mmNotificationClient == nil {
//		if err := sf.registerDefaultDeviceChangeCallback(); err != nil {
//			return nil, fmt.Errorf("register default device change callback: %w", err)
//		}
//	}
//
//	// get the currently active default output and input devices.
//	// please note that this can return a nil defaultInputEndpoint, in case there are no input devices connected.
//	// you must check it for non-nil
//	defaultOutputEndpoint, err := sf.getDefaultAudioEndpoints()
//	if err != nil {
//		return nil, fmt.Errorf("get default audio endpoints: %w", err)
//	}
//	defer defaultOutputEndpoint.Release()
//
//	// get the master output session
//	var audioEndpointVolume *wca.IAudioEndpointVolume
//	if err := defaultOutputEndpoint.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &audioEndpointVolume); err != nil {
//		return nil, fmt.Errorf("activate master session: %w", err)
//	}
//
//	//sessions = append(sessions, sf.masterOut)
//	//
//	//// get the master input session, if a default input device exists
//	//if defaultInputEndpoint != nil {
//	//	sf.masterIn, err = sf.getMasterSession(defaultInputEndpoint, inputSessionName, inputSessionName)
//	//	if err != nil {
//	//		sf.logger.Warnw("Failed to get master audio input session", "error", err)
//	//		return nil, fmt.Errorf("get master audio input session: %w", err)
//	//	}
//	//
//	//	sessions = append(sessions, sf.masterIn)
//	//}
//
//	fmt.Println("WE GOOD")
//
//	return nil, nil
//}
//
//func (sf *source) registerDefaultDeviceChangeCallback() error {
//	sf.mmNotificationClient = &wca.IMMNotificationClient{}
//
//	sf.mmNotificationClient = wca.NewIMMNotificationClient(wca.IMMNotificationClientCallback{
//		OnDefaultDeviceChanged: sf.defaultDeviceChangedCallback,
//		OnDeviceAdded: func(pwstrDeviceId string) error {
//			return nil
//		},
//		OnDeviceRemoved: func(pwstrDeviceId string) error {
//			return nil
//		},
//		OnDeviceStateChanged: func(pwstrDeviceId string, dwNewState uint64) error {
//			return nil
//		},
//		OnPropertyValueChanged: func(pwstrDeviceId string, dwNewValue uint64) error {
//			return nil
//		},
//	})
//
//	if err := sf.mmDeviceEnumerator.RegisterEndpointNotificationCallback(sf.mmNotificationClient); err != nil {
//		return fmt.Errorf("call RegisterEndpointNotificationCallback: %w", err)
//	}
//
//	return nil
//}
//
//const minDefaultDeviceChangeThreshold = 100 * time.Millisecond
//
//func (sf *source) defaultDeviceChangedCallback(
//	flow wca.EDataFlow, role wca.ERole, pwstrDeviceId string,
//) error {
//
//	fmt.Println("defaultDeviceChangedCallback")
//
//	// filter out calls that happen in rapid succession
//	now := time.Now()
//
//	if sf.lastDefaultDeviceChange.Add(minDefaultDeviceChangeThreshold).After(now) {
//		return nil
//	}
//
//	sf.lastDefaultDeviceChange = now
//
//	fmt.Println("DEVICE CHANGED!")
//
//	sf.stop()
//
//	err := sf.start()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	//sf.logger.Debug("Default audio device changed, marking master sessions as stale")
//	//if sf.masterOut != nil {
//	//	sf.masterOut.markAsStale()
//	//}
//	//
//	//if sf.masterIn != nil {
//	//	sf.masterIn.markAsStale()
//	//}
//
//	return nil
//}
//
//func (sf *source) getDefaultAudioEndpoints() (*wca.IMMDevice, error) {
//
//	// get the default audio endpoints as IMMDevice instances
//	var mmOutDevice *wca.IMMDevice
//
//	if err := sf.mmDeviceEnumerator.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmOutDevice); err != nil {
//		return nil, fmt.Errorf("call GetDefaultAudioEndpoint (out): %w", err)
//	}
//
//	return mmOutDevice, nil
//}
//
//func (sf *source) getDeviceEnumerator() error {
//
//	// get the IMMDeviceEnumerator (only once)
//	if sf.mmDeviceEnumerator == nil {
//		if err := wca.CoCreateInstance(
//			wca.CLSID_MMDeviceEnumerator,
//			0,
//			wca.CLSCTX_ALL,
//			wca.IID_IMMDeviceEnumerator,
//			&sf.mmDeviceEnumerator,
//		); err != nil {
//			return fmt.Errorf("call CoCreateInstance: %w", err)
//		}
//	}
//
//	return nil
//}
//
//func (s *source) stop() {
//	if s.cancel == nil {
//		return
//	}
//
//	for _, in := range s.inputs {
//		in.close()
//		//s.reg.RemoveInput(in.id)
//	}
//
//	s.cancel()
//	s.wg.Wait()
//	s.cancel = nil
//}
//
//func (s *source) start() error {
//	// get list of devices
//	var deviceCollection *wca.IMMDeviceCollection
//
//	if err := s.mmDeviceEnumerator.EnumAudioEndpoints(wca.ERender, wca.DEVICE_STATE_ACTIVE, &deviceCollection); err != nil {
//		return fmt.Errorf("enumerate active audio endpoints: %w", err)
//	}
//
//	// check how many devices there are
//	var deviceCount uint32
//	if err := deviceCollection.GetCount(&deviceCount); err != nil {
//		return fmt.Errorf("get device count from device collection: %w", err)
//	}
//
//	newInputs := map[uuid.UUID]*input{}
//
//	// for each device:
//	for deviceIdx := uint32(0); deviceIdx < deviceCount; deviceIdx++ {
//		fmt.Println("PROCESS DEVICE", deviceIdx)
//
//		in, err := newInput(s.id, deviceIdx, deviceCollection)
//		if err != nil {
//			return err
//		}
//
//		newInputs[in.id] = in
//
//		//// get its IMMDevice instance
//		//var endpoint *wca.IMMDevice
//		//
//		//if err := deviceCollection.Item(deviceIdx, &endpoint); err != nil {
//		//	return fmt.Errorf("get device %d from device collection: %w", deviceIdx, err)
//		//}
//		//defer endpoint.close()
//		//
//		/////*
//		//
//		////// get its IMMEndpoint instance to figure out if it's an output device (and we need to enumerate its process sessions later)
//		////dispatch, err := endpoint.QueryInterface(wca.IID_IMMEndpoint)
//		////if err != nil {
//		////	return fmt.Errorf("query device %d IMMEndpoint: %w", deviceIdx, err)
//		////}
//		//
//		//// get the device's property store
//		//var propertyStore *wca.IPropertyStore
//		//
//		//if err := endpoint.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
//		//	return fmt.Errorf("open endpoint %d property store: %w", deviceIdx, err)
//		//}
//		//defer propertyStore.close()
//		//
//		//// query the property store for the device's description and friendly name
//		//value := &wca.PROPVARIANT{}
//		//
//		//// device uuid
//		//if err := propertyStore.GetValue(&wca.PKEY_AudioEndpoint_GUID, value); err != nil {
//		//	return fmt.Errorf("get device %d description: %w", deviceIdx, err)
//		//}
//		//endpointUuid := strings.ToLower(value.String())
//		//
//		//// device description i.e. "Headphones"
//		//if err := propertyStore.GetValue(&wca.PKEY_Device_DeviceDesc, value); err != nil {
//		//	return fmt.Errorf("get device %d description: %w", deviceIdx, err)
//		//}
//		//endpointDescription := strings.ToLower(value.String())
//		//
//		//// device friendly name i.e. "Headphones (Realtek Audio)"
//		//if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value); err != nil {
//		//	return fmt.Errorf("get device %d friendly name: %w", deviceIdx, err)
//		//}
//		//endpointFriendlyName := value.String()
//		//
//		////// receive a useful object instead of our dispatch
//		////endpointType := (*wca.IMMEndpoint)(unsafe.Pointer(dispatch))
//		////defer endpointType.close()
//		//
//		////var dataFlow uint32
//		////if err := endpointType.GetDataFlow(&dataFlow); err != nil {
//		////	return fmt.Errorf("get device %d data flow: %w", deviceIdx, err)
//		////}
//		//
//		//fmt.Println("Enumerated device info",
//		//	"deviceIdx", deviceIdx,
//		//	"deviceDescription", endpointDescription,
//		//	"deviceFriendlyName", endpointFriendlyName,
//		//	"uuid", endpointUuid,
//		//	//"dataFlow", dataFlow,
//		//)
//		// if the device is an output device, enumerate and add its per-process audio sessions
//		//if dataFlow == wca.ERender {
//		//	fmt.Println("enumerateAndAddProcessSessions")
//		//	//if err := sf.enumerateAndAddProcessSessions(endpoint, endpointFriendlyName, sessions); err != nil {
//		//	//	return fmt.Errorf("enumerate and add device %d process sessions: %w", deviceIdx, err)
//		//	//}
//		//}
//
//		//go func() {
//		//	return
//		//
//		//	err := s.captureDev(deviceIdx, deviceCollection)
//		//	if err != nil {
//		//		fmt.Println("error during capture", err)
//		//	}
//		//}()
//
//		// for all devices (both input and output), add a named "master" session that can be addressed
//		// by using the device's friendly name (as appears when the user left-clicks the speaker icon in the tray)
//		//newSession, err := sf.getMasterSession(endpoint,
//		//	endpointFriendlyName,
//		//	fmt.Sprintf(deviceSessionFormat, endpointDescription))
//		//
//		//if err != nil {
//		//	sf.logger.Warnw("Failed to get master session for device",
//		//		"deviceIdx", deviceIdx,
//		//		"error", err)
//		//
//		//	return fmt.Errorf("get device %d master session: %w", deviceIdx, err)
//		//}
//		//
//		//// add it to our slice
//		//*sessions = append(*sessions, newSession)
//	}
//
//	fmt.Println("old", lo.Keys(s.inputs))
//	fmt.Println("new", lo.Keys(newInputs))
//
//	for id, in := range s.inputs {
//		//in.close()
//
//		if _, ok := newInputs[id]; !ok {
//			s.reg.RemoveInput(in.id)
//			delete(s.inputs, id)
//		}
//	}
//
//	for _, in := range newInputs {
//		if _, ok := s.inputs[in.id]; !ok {
//			s.reg.AddInput(in)
//		}
//
//		s.inputs[in.id] = in
//	}
//
//	return nil
//}
//
////func (sf *source) enumerateAndAddProcessSessions(
////	endpoint *wca.IMMDevice,
////	endpointFriendlyName string,
////	sessions *[]Session,
////) error {
////	fmt.Println("Enumerating and adding process sessions for audio output device",
////		"deviceFriendlyName", endpointFriendlyName)
////
////	// query the given IMMDevice's IAudioSessionManager2 interface
////	var audioSessionManager2 *wca.IAudioSessionManager2
////
////	if err := endpoint.Activate(
////		wca.IID_IAudioSessionManager2,
////		wca.CLSCTX_ALL,
////		nil,
////		&audioSessionManager2,
////	); err != nil {
////		return fmt.Errorf("activate endpoint: %w", err)
////	}
////	defer audioSessionManager2.close()
////
////	// get its IAudioSessionEnumerator
////	var sessionEnumerator *wca.IAudioSessionEnumerator
////
////	if err := audioSessionManager2.GetSessionEnumerator(&sessionEnumerator); err != nil {
////		return err
////	}
////	defer sessionEnumerator.close()
////
////	// check how many audio sessions there are
////	var sessionCount int
////
////	if err := sessionEnumerator.GetCount(&sessionCount); err != nil {
////		return fmt.Errorf("get session count: %w", err)
////	}
////
////	fmt.Println("Got session count from session enumerator", "count", sessionCount)
////
////	// for each session:
////	for sessionIdx := 0; sessionIdx < sessionCount; sessionIdx++ {
////
////		// get the IAudioSessionControl
////		var audioSessionControl *wca.IAudioSessionControl
////		if err := sessionEnumerator.GetSession(sessionIdx, &audioSessionControl); err != nil {
////			return fmt.Errorf("get session %d from enumerator: %w", sessionIdx, err)
////		}
////
////		// query its IAudioSessionControl2
////		dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
////		if err != nil {
////			return fmt.Errorf("query session %d IAudioSessionControl2: %w", sessionIdx, err)
////		}
////
////		// we no longer need the IAudioSessionControl, release it
////		audioSessionControl.close()
////
////		// receive a useful object instead of our dispatch
////		audioSessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))
////
////		var pid uint32
////
////		// get the session's PID
////		if err := audioSessionControl2.GetProcessId(&pid); err != nil {
////
////			// if this is the system sounds session, GetProcessId will error with an undocumented
////			// AUDCLNT_S_NO_CURRENT_PROCESS (0x889000D) - this is fine, we actually want to treat it a bit differently
////			// The first part of this condition will be true if the call to IsSystemSoundsSession fails
////			// The second part will be true if the original error mesage from GetProcessId doesn't contain this magical
////			// error code (in decimal format).
////			isSystemSoundsErr := audioSessionControl2.IsSystemSoundsSession()
////			if isSystemSoundsErr != nil && !strings.Contains(err.Error(), "143196173") {
////
////				// of course, if it's not the system sounds session, we got a problem
////				return fmt.Errorf("query session %d pid: %w", sessionIdx, err)
////			}
////
////			// update 2020/08/31: this is also the exact case for UWP applications, so we should no longer override the PID.
////			// it will successfully update whenever we call GetProcessId for e.g. Video.UI.exe, despite the error being non-nil.
////		}
////
////		// get its ISimpleAudioVolume
////		dispatch, err = audioSessionControl2.QueryInterface(wca.IID_ISimpleAudioVolume)
////		if err != nil {
////			return fmt.Errorf("query session %d ISimpleAudioVolume: %w", sessionIdx, err)
////		}
////
////		// make it useful, again
////		simpleAudioVolume := (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch))
////
////		// create the deej session object
////		newSession, err := newWCASession(sf.sessionLogger, audioSessionControl2, simpleAudioVolume, pid, sf.eventCtx)
////		if err != nil {
////
////			// this could just mean this process is already closed by now, and the session will be cleaned up later by the OS
////			if !errors.Is(err, errNoSuchProcess) {
////				sf.logger.Warnw("Failed to create new WCA session instance",
////					"error", err,
////					"sessionIdx", sessionIdx)
////
////				return fmt.Errorf("create wca session for session %d: %w", sessionIdx, err)
////			}
////
////			// in this case, log it and release the session's handles, then skip to the next one
////			sf.logger.Debugw("Process already exited, skipping session and releasing handles", "pid", pid)
////
////			audioSessionControl2.close()
////			simpleAudioVolume.close()
////
////			continue
////		}
////
////		// add it to our slice
////		*sessions = append(*sessions, newSession)
////	}
////
////	return nil
////}
//
//func (s *source) startOld() error {
//	ctx := context.Background()
//
//	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
//	if err != nil {
//		return err
//	}
//	defer ole.CoUninitialize()
//
//	var mmde *wca.IMMDeviceEnumerator
//	err = wca.CoCreateInstance(
//		wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL,
//		wca.IID_IMMDeviceEnumerator, &mmde,
//	)
//	if err != nil {
//		return err
//	}
//	defer mmde.Release()
//
//	////////////////////////////
//
//	var deviceCollection *wca.IMMDeviceCollection
//
//	err = mmde.EnumAudioEndpoints(wca.ERender, wca.DEVICE_STATE_ACTIVE, &deviceCollection)
//	if err != nil {
//		return fmt.Errorf("enumerate active audio endpoints: %w", err)
//	}
//	defer deviceCollection.Release()
//
//	var deviceCount uint32
//	if err := deviceCollection.GetCount(&deviceCount); err != nil {
//		return fmt.Errorf("get device count from device collection: %w", err)
//	}
//
//	fmt.Println(deviceCount)
//
//	////////////////////////////
//
//	fmt.Println("exit")
//	return nil
//
//	////////////////////////////
//
//	var mmd *wca.IMMDevice
//	err = mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd)
//	if err != nil {
//		return err
//	}
//	defer mmd.Release()
//
//	var ps *wca.IPropertyStore
//	err = mmd.OpenPropertyStore(wca.STGM_READ, &ps)
//	if err != nil {
//		return err
//	}
//	defer ps.Release()
//
//	var audioCli *wca.IAudioClient
//	err = mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioCli)
//	if err != nil {
//		return err
//	}
//	defer audioCli.Release()
//
//	var ami *wcaami.IAudioMeterInformation
//	err = mmd.Activate(
//		wca.IID_IAudioMeterInformation, wca.CLSCTX_ALL, nil, &ami,
//	)
//	if err != nil {
//		return err
//	}
//	defer ami.Release()
//
//	var wfx *wca.WAVEFORMATEX
//	err = audioCli.GetMixFormat(&wfx)
//	if err != nil {
//		return err
//	}
//	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))
//
//	var defaultPeriod wca.REFERENCE_TIME
//	var minimumPeriod wca.REFERENCE_TIME
//	var latency time.Duration
//
//	err = audioCli.GetDevicePeriod(&defaultPeriod, &minimumPeriod)
//	if err != nil {
//		return err
//	}
//	latency = time.Duration(int(defaultPeriod) * 100)
//
//	wfx.NChannels = 2 // force stereo
//	wfx.WFormatTag = 1
//	wfx.WBitsPerSample = 32
//	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
//	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
//	wfx.CbSize = 0
//
//	err = audioCli.Initialize(
//		wca.AUDCLNT_SHAREMODE_SHARED,
//		wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK,
//		defaultPeriod, 0, wfx, nil,
//	)
//	if err != nil {
//		return err
//	}
//
//	audioReadyEvent := wca.CreateEventExA(
//		0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE,
//	)
//	defer func() {
//		_ = wca.CloseHandle(audioReadyEvent)
//	}()
//
//	err = audioCli.SetEventHandle(audioReadyEvent)
//	if err != nil {
//		return err
//	}
//
//	var bufferFrameSize uint32
//	err = audioCli.GetBufferSize(&bufferFrameSize)
//	if err != nil {
//		return err
//	}
//
//	var acc *wca.IAudioCaptureClient
//	err = audioCli.GetService(wca.IID_IAudioCaptureClient, &acc)
//	if err != nil {
//		return err
//	}
//	defer acc.Release()
//
//	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
//	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
//	fmt.Printf("Channels: %d\n", wfx.NChannels)
//
//	fmt.Println("Default period: ", defaultPeriod)
//	fmt.Println("Minimum period: ", minimumPeriod)
//	fmt.Println("Latency: ", latency)
//
//	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)
//
//	err = audioCli.Start()
//	if err != nil {
//		return err
//	}
//
//	var offset int
//	var b *byte
//	var data *byte
//	var availableFrameSize uint32
//	var flags uint32
//	var devicePosition uint64
//	var qcpPosition uint64
//
//	errorChan := make(chan error, 1)
//
//	var isCapturing = true
//
//loop:
//	for {
//		if !isCapturing {
//			close(errorChan)
//			break
//		}
//		go func() {
//			errorChan <- watchEvent(ctx, audioReadyEvent)
//		}()
//
//		select {
//		case <-ctx.Done():
//			isCapturing = false
//			<-errorChan
//			break loop
//		case err := <-errorChan:
//			if err != nil {
//				isCapturing = false
//				break
//			}
//			err = acc.GetBuffer(
//				&data, &availableFrameSize, &flags,
//				&devicePosition, &qcpPosition,
//			)
//
//			if err != nil {
//				continue
//			}
//
//			if availableFrameSize == 0 {
//				continue
//			}
//
//			start := unsafe.Pointer(data)
//			if start == nil {
//				continue
//			}
//
//			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
//			buf := make([]byte, lim)
//
//			for n := 0; n < lim; n++ {
//				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
//				buf[n] = *b
//			}
//
//			// close the buffer as soon as we extract the audio Samples
//			err = acc.ReleaseBuffer(availableFrameSize)
//			if err != nil {
//				return fmt.Errorf("failed to release buffer: %w", err)
//			}
//
//			offset += lim
//
//			samples := make([]float64, len(buf)/4)
//			for i := 0; i < len(buf); i += 4 {
//				v := float64(readInt32(buf[i : i+4]))
//				samples = append(samples, v)
//			}
//
//			// TODO: calculate impact of this call
//			var peak float32
//			err = ami.GetPeakValue(&peak)
//			if err != nil {
//				continue
//			}
//			//Peak = 1
//
//			// Dispatch the received frame for processing. If the work queue
//			// is full, this will block until c previous frame is processed.
//			s.frames <- frame{
//				Samples: samples,
//				Peak:    float64(peak),
//			}
//		}
//	}
//
//	err = audioCli.Stop()
//	if err != nil {
//		return fmt.Errorf("failed to stop audio client: %w", err)
//	}
//
//	return nil
//}
//
//func (s *source) captureDev(deviceIdx uint32, deviceCollection *wca.IMMDeviceCollection) error {
//
//	// get its IMMDevice instance
//	var mmd *wca.IMMDevice
//
//	if err := deviceCollection.Item(deviceIdx, &mmd); err != nil {
//		return fmt.Errorf("get device %d from device collection: %w", deviceIdx, err)
//	}
//	defer mmd.Release()
//
//	///*
//
//	//// get its IMMEndpoint instance to figure out if it's an output device (and we need to enumerate its process sessions later)
//	//dispatch, err := mmd.QueryInterface(wca.IID_IMMEndpoint)
//	//if err != nil {
//	//	return fmt.Errorf("query device %d IMMEndpoint: %w", deviceIdx, err)
//	//}
//
//	// get the device's property store
//	var propertyStore *wca.IPropertyStore
//
//	if err := mmd.OpenPropertyStore(wca.STGM_READ, &propertyStore); err != nil {
//		return fmt.Errorf("open mmd %d property store: %w", deviceIdx, err)
//	}
//	defer propertyStore.Release()
//
//	// query the property store for the device's description and friendly name
//	value := &wca.PROPVARIANT{}
//
//	// device uuid
//	if err := propertyStore.GetValue(&wca.PKEY_AudioEndpoint_GUID, value); err != nil {
//		return fmt.Errorf("get device %d description: %w", deviceIdx, err)
//	}
//	endpointUuid := strings.ToLower(value.String())
//
//	// device description i.e. "Headphones"
//	if err := propertyStore.GetValue(&wca.PKEY_Device_DeviceDesc, value); err != nil {
//		return fmt.Errorf("get device %d description: %w", deviceIdx, err)
//	}
//	endpointDescription := strings.ToLower(value.String())
//
//	// device friendly name i.e. "Headphones (Realtek Audio)"
//	if err := propertyStore.GetValue(&wca.PKEY_Device_FriendlyName, value); err != nil {
//		return fmt.Errorf("get device %d friendly name: %w", deviceIdx, err)
//	}
//	endpointFriendlyName := value.String()
//
//	//// receive a useful object instead of our dispatch
//	//endpointType := (*wca.IMMEndpoint)(unsafe.Pointer(dispatch))
//	//defer endpointType.close()
//
//	//var dataFlow uint32
//	//if err := endpointType.GetDataFlow(&dataFlow); err != nil {
//	//	return fmt.Errorf("get device %d data flow: %w", deviceIdx, err)
//	//}
//
//	fmt.Println("Enumerated device info",
//		"deviceIdx", deviceIdx,
//		"deviceDescription", endpointDescription,
//		"deviceFriendlyName", endpointFriendlyName,
//		"uuid", endpointUuid,
//		//"dataFlow", dataFlow,
//	)
//
//	// setup input
//
//	ctx := context.Background()
//
//	var audioCli *wca.IAudioClient
//	err := mmd.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioCli)
//	if err != nil {
//		return err
//	}
//	defer audioCli.Release()
//
//	//var ami *wcaami.IAudioMeterInformation
//	//err = mmd.Activate(
//	//	wca.IID_IAudioMeterInformation, wca.CLSCTX_ALL, nil, &ami,
//	//)
//	//if err != nil {
//	//	return err
//	//}
//	//defer ami.close()
//
//	var wfx *wca.WAVEFORMATEX
//	err = audioCli.GetMixFormat(&wfx)
//	if err != nil {
//		return err
//	}
//	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))
//
//	var defaultPeriod wca.REFERENCE_TIME
//	var minimumPeriod wca.REFERENCE_TIME
//	var latency time.Duration
//
//	err = audioCli.GetDevicePeriod(&defaultPeriod, &minimumPeriod)
//	if err != nil {
//		return err
//	}
//	latency = time.Duration(int(defaultPeriod) * 100)
//
//	wfx.NChannels = 2 // force stereo
//	wfx.WFormatTag = 1
//	wfx.WBitsPerSample = 32
//	wfx.NBlockAlign = (wfx.WBitsPerSample / 8) * wfx.NChannels
//	wfx.NAvgBytesPerSec = wfx.NSamplesPerSec * uint32(wfx.NBlockAlign)
//	wfx.CbSize = 0
//
//	err = audioCli.Initialize(
//		wca.AUDCLNT_SHAREMODE_SHARED,
//		wca.AUDCLNT_STREAMFLAGS_EVENTCALLBACK|wca.AUDCLNT_STREAMFLAGS_LOOPBACK,
//		defaultPeriod, 0, wfx, nil,
//	)
//	if err != nil {
//		return err
//	}
//
//	audioReadyEvent := wca.CreateEventExA(
//		0, 0, 0, wca.EVENT_MODIFY_STATE|wca.SYNCHRONIZE,
//	)
//	defer func() {
//		_ = wca.CloseHandle(audioReadyEvent)
//	}()
//
//	err = audioCli.SetEventHandle(audioReadyEvent)
//	if err != nil {
//		return err
//	}
//
//	var bufferFrameSize uint32
//	err = audioCli.GetBufferSize(&bufferFrameSize)
//	if err != nil {
//		return err
//	}
//
//	var acc *wca.IAudioCaptureClient
//	err = audioCli.GetService(wca.IID_IAudioCaptureClient, &acc)
//	if err != nil {
//		return err
//	}
//	defer acc.Release()
//
//	fmt.Printf("Format: PCM %d bit signed integer\n", wfx.WBitsPerSample)
//	fmt.Printf("Rate: %d Hz\n", wfx.NSamplesPerSec)
//	fmt.Printf("Channels: %d\n", wfx.NChannels)
//
//	fmt.Println("Default period: ", defaultPeriod)
//	fmt.Println("Minimum period: ", minimumPeriod)
//	fmt.Println("Latency: ", latency)
//
//	fmt.Printf("Allocated buffer size: %d\n", bufferFrameSize)
//
//	err = audioCli.Start()
//	if err != nil {
//		return err
//	}
//
//	var offset int
//	var b *byte
//	var data *byte
//	var availableFrameSize uint32
//	var flags uint32
//	var devicePosition uint64
//	var qcpPosition uint64
//
//	errorChan := make(chan error, 1)
//
//	var isCapturing = true
//
//loop:
//	for {
//		if !isCapturing {
//			close(errorChan)
//			break
//		}
//		go func() {
//			errorChan <- watchEvent(ctx, audioReadyEvent)
//		}()
//
//		select {
//		case <-ctx.Done():
//			isCapturing = false
//			<-errorChan
//			break loop
//		case err := <-errorChan:
//			if err != nil {
//				isCapturing = false
//				break
//			}
//			err = acc.GetBuffer(
//				&data, &availableFrameSize, &flags,
//				&devicePosition, &qcpPosition,
//			)
//
//			if err != nil {
//				continue
//			}
//
//			if availableFrameSize == 0 {
//				continue
//			}
//
//			start := unsafe.Pointer(data)
//			if start == nil {
//				continue
//			}
//
//			lim := int(availableFrameSize) * int(wfx.NBlockAlign)
//			buf := make([]byte, lim)
//
//			for n := 0; n < lim; n++ {
//				b = (*byte)(unsafe.Pointer(uintptr(start) + uintptr(n)))
//				buf[n] = *b
//			}
//
//			// close the buffer as soon as we extract the audio Samples
//			err = acc.ReleaseBuffer(availableFrameSize)
//			if err != nil {
//				return fmt.Errorf("failed to release buffer: %w", err)
//			}
//
//			offset += lim
//
//			samples := make([]float64, len(buf)/4)
//			for i := 0; i < len(buf); i += 4 {
//				v := float64(readInt32(buf[i : i+4]))
//				samples = append(samples, v)
//			}
//
//			// TODO: calculate impact of this call
//			//var peak float32
//			//err = ami.GetPeakValue(&peak)
//			//if err != nil {
//			//	continue
//			//}
//			//Peak = 1
//
//			// Dispatch the received frame for processing. If the work queue
//			// is full, this will block until c previous frame is processed.
//
//			fmt.Println(endpointUuid, "frame")
//
//			select {
//			case s.frames <- frame{
//				Samples: samples,
//				//Peak:    float64(peak),
//			}:
//			default:
//			}
//		}
//	}
//
//	err = audioCli.Stop()
//	if err != nil {
//		return fmt.Errorf("failed to stop audio client: %w", err)
//	}
//
//	return nil
//}
//
//func watchEvent(ctx context.Context, event uintptr) (err error) {
//	errorChan := make(chan error, 1)
//	go func() {
//		errorChan <- eventEmitter(event)
//	}()
//	select {
//	case err = <-errorChan:
//		close(errorChan)
//		return
//	case <-ctx.Done():
//		err = ctx.Err()
//		return
//	}
//}
//
//func eventEmitter(event uintptr) (err error) {
//	dw := wca.WaitForSingleObject(event, wca.INFINITE)
//	if dw != 0 {
//		return fmt.Errorf("failed to watch event")
//	}
//	return nil
//}
//
//// readInt32 reads a signed integer from a byte slice. only a slice with len(4)
//// should be passed. equivalent to int32(binary.LittleEndian.Uint32(b))
//func readInt32(b []byte) int32 {
//	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
//}
