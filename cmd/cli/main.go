package main

import (
	"fmt"
	"os"
)

func main() {
	root := Root()

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
