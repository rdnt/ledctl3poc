#!/bin/bash
go build -o ./build/ledctl ./cmd/cli
./build/ledctl completion bash > ~/.bash_completion
. ~/.bash_completion
