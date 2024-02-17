package main

import (
	"fmt"
	"testing"

	"ledctl3/shlex"
)

func TestTui(_ *testing.T) {
	args, err := shlex.Split("test \"unquo")
	fmt.Println(args, err)
	//t := shlex.NewTokenizer(strings.NewReader())
	//for {
	//	token, err := t.Next()
	//	if err != nil {
	//		fmt.Println("ERR", token)
	//		break
	//	}
	//	fmt.Println("TOK", token)
	//}

}
