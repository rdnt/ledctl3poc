package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"ledctl3/event"
	"ledctl3/internal/registry"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
	"ledctl3/pkg/uuid"
)

type sh struct {
	mux sync.Mutex
}

func (s *sh) SetState(state registry.State) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("./registry.json", b, 0644)
}

func (s *sh) GetState() (registry.State, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	b, err := os.ReadFile("./registry.json")
	if err != nil {
		return registry.State{}, err
	}

	var state registry.State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return registry.State{}, err
	}

	return state, nil
}

func main() {
	s := netserver.New[event.Event](1337, event.Codec)

	sh := &sh{}
	reg := registry.New(sh, func(addr string, e event.Event) error {
		return s.Write(addr, e)
	})

	s.SetMessageHandler(func(addr string, e event.Event) {
		reg.ProcessEvent(addr, e)
	})

	s.SetConnectHandler(func(addr string) {
		//fmt.Println("device connected")
	})

	s.SetDisconnectHandler(func(addr string) {
		reg.ProcessEvent(addr, event.Disconnect{})
	})

	time.Sleep(1 * time.Second)
	fmt.Println("registry started")

	err := s.Start()
	if err != nil {
		panic(err)
	}

	mdnsServer, err := mdns.NewServer("registry", 1337)
	if err != nil {
		panic(err)
	}

	err = mdnsServer.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(5 * time.Second)

		fmt.Println("Updating driver config!")
		err = reg.SetDeviceConfig(uuid.MustParse("faf5dc1b-0001-4654-bec8-9eecc18f38a0"), uuid.MustParse("f8b279f7-a39c-43df-87c8-9fe7ffadc51d"), []byte(`
{
  "outputs": [
    {
      "id": "0000aaaa-0000-0000-0000-000000000000",
      "count": 1,
      "offset": 0
    },
    {
      "id": "0000bbbb-0000-0000-0000-000000000000",
      "count": 1,
      "offset": 1
    },
    {
      "id": "0000cccc-0000-0000-0000-000000000000",
      "count": 1,
      "offset": 2
    }
  ]
}
		`))
		if err != nil {
			fmt.Println("ERR!", err)
			//panic(err)
		}
	}()

	//go func() {
	//	time.Sleep(10 * time.Second)
	//
	//	fmt.Println("Activating profile!")
	//	err = reg.EnableProfile(uuid.MustParse("ffffffff-2e2d-4470-b9ab-c78786bf5667"))
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	_, err = reg.CreateProfile("custom", []registry.ProfileSource{
	//		{
	//			OutputId: uuid.MustParse("55555555-dca5-430b-971c-fbe5b9112bfe"),
	//			Inputs: []registry.ProfileInput{
	//				{
	//					OutputId: uuid.MustParse("22222222-b301-47d6-b289-2a4c3327962a"),
	//					Sinks: []registry.ProfileSink{
	//						{
	//							OutputId: uuid.MustParse("55555555-dca5-430b-971c-fbe5b9112bfe"),
	//							Outputs: []registry.ProfileOutput{
	//								{
	//									OutputId:            uuid.MustParse("88888888-6b50-4789-b635-16237d268efa"),
	//									InputConfigId: uuid.Nil,
	//								},
	//							},
	//						},
	//					},
	//				},
	//				{
	//					OutputId: uuid.MustParse("33333333-e72d-470e-a343-5c2cc2f1746f"),
	//					Sinks: []registry.ProfileSink{
	//						{
	//							OutputId: uuid.MustParse("55555555-dca5-430b-971c-fbe5b9112bfe"),
	//							Outputs: []registry.ProfileOutput{
	//								{
	//									OutputId:            uuid.MustParse("88888888-6b50-4789-b635-16237d268efa"),
	//									InputConfigId: uuid.Nil,
	//								},
	//							},
	//						},
	//					},
	//				},
	//			},
	//		},
	//	})
	//	if err != nil {
	//		fmt.Println("error adding profile:", err)
	//	}
	//}()

	select {}
}
