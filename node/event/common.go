package event

type InputType string

const (
	InputTypeDefault       InputType = "default"
	InputTypeScreenCapture InputType = "screen_capture"
	InputTypeAudioCapture  InputType = "audio_capture"
)
