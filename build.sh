#!/bin/bash
go build -o ./build/ledctl.exe ./cmd/cli
./build/ledctl.exe completion bash > ~/.bash_completion
. ~/.bash_completion
