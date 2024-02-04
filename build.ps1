go build -o ./build/ledctl.exe ./cmd/registry
./build/ledctl.exe completion powershell | Out-String | Invoke-Expression
./build/ledctl.exe completion powershell
