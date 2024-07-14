package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	aud, err := newSessionFinder()
	if err != nil {
		panic(err)
	}

	fmt.Println("starting")

	aud.start()

	fmt.Println("started")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("stopping")

	aud.stop()

	fmt.Println("done")
}
