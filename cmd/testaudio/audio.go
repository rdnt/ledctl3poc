package main

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
)

type source struct {
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

func newSessionFinder() (*source, error) {
	sf := &source{
		inputs:       make(map[uuid.UUID]*input),
		updateVolume: make(chan bool, 1),
	}

	fmt.Println("debug: Created WCA session finder instance")

	return sf, nil
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
	master, err := newInput("main", "main", defaultOutputEndpoint)
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

		in, err := newInput(endpointUuid.String(), endpointFriendlyName, endpoint)
		if err != nil {
			fmt.Println("warn: Failed to get main session for device",
				"deviceIdx", deviceIdx,
				"error", err)

			return fmt.Errorf("get device %d main session: %w", deviceIdx, err)
		}

		s.inputs[endpointUuid] = in
	}

	return nil
}
