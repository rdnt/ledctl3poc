go build -o ./build/ledctl ./cmd/cli
./build/ledctl completion powershell | Out-String | Invoke-Expression
