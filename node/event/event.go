package event

const (
	TypeAuthenticate       = "Authenticate"
	TypeCapabilities       = "Capabilities"
	TypeData               = "Data"
	TypeSetSourceActive    = "SetSourceActive"
	TypeSetInputActive     = "SetInputActive"
	TypeInputConnected     = "InputConnected"
	TypeInputDisconnected  = "InputDisconnected"
	TypeOutputConnected    = "OutputConnected"
	TypeOutputDisconnected = "OutputDisconnected"
	TypeSetSourceConfig    = "SetSourceConfig"
	TypeSetSinkConfig      = "SetSinkConfig"

	TypeResponse = "CommandResponse"

	TypeSetSourceConfigCommand = "SetSourceConfigCommand"
	TypeSetSinkConfigCommand   = "SetSinkConfigCommand"
)
