package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledctl3/cmd/registry/state"
	"ledctl3/node/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver"
	"ledctl3/pkg/uuid"
	"ledctl3/registry"
)

func main() {
	s := netserver.New[event.Event](1337, event.NewJSONCodec())

	sh := state.NewHolder()
	reg := registry.New(sh, func(addr string, e event.Event) error {
		return s.Write(addr, e)
	})

	s.SetMessageHandler(func(addr string, e event.Event) {
		err := reg.ProcessEvent(addr, e)
		if err != nil {
			fmt.Println("error processing event:", err)
		}
	})

	s.SetRequestHandler(func(addr string, e event.Event, respond func(event.Event) error) {
		//ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		//defer cancel()

		err := reg.ProcessEvent(addr, e)
		if err != nil {
			fmt.Println("error processing event:", err)
		}

		var errorResponse *string
		if err != nil {
			errStr := err.Error()
			errorResponse = &errStr
		}

		err2 := respond(event.Response{
			Success: err == nil,
			Error:   errorResponse,
		})

		if err2 != nil {
			fmt.Println("error responding to request:", err2)
		}
	})

	s.SetConnectHandler(reg.HandleConnected)

	s.SetDisconnectHandler(reg.HandleDisconnected)

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
		err = reg.SetSinkConfig(uuid.MustParse("ffffffff-0000-0000-0000-000000000000"), uuid.MustParse("2b599945-732d-4a1c-afc2-4ffd07c4131b"), []byte(`
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

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	sh.Stop()
}
