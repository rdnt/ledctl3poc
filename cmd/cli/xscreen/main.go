package main

import (
	"os"
	"os/signal"

	"github.com/google/uuid"

	"ledctl3/node"
	_ "ledctl3/node/screen2"
)

func main() {
	_, err := node.New(
		node.Config{
			Id: uuid.Nil,
		})
	if err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
