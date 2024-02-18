go build -o ./build/ledctl.exe ./cmd/cli
./build/ledctl.exe completion powershell | Out-String | Invoke-Expression

#source <(./build/ledctl.exe completion bash)
