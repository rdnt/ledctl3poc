package event

import (
	"ledctl3/pkg/uuid"
)

type Type string

const (
	SetIdle   Type = "setIdle"
	SetActive Type = "setActive"
	SetLeds   Type = "setLeds"
)

type EventIface interface {
	Type() Type
	DeviceId() uuid.UUID
	//DeviceId() uuid.UUID
}

const (
	Connect             Type = "connect"
	ListCapabilities    Type = "listCapabilities"
	Capabilities        Type = "capabilities"
	SetSinkActive       Type = "setActive"
	SetSourceIdle       Type = "setIdle"
	SetSourceActive     Type = "setActive"
	Data                Type = "data"
	SetInputConfig      Type = "setInputConfig"
	AssistedSetup       Type = "assistedSetup"
	AssistedSetupConfig Type = "assistedSetupConfig"
)

type Event struct {
	Type  Type      `json:"event"`
	DevId uuid.UUID `json:"deviceId"`
}

func (e Event) DeviceId() uuid.UUID {
	return e.DevId
}

//type Type interface {
//	Type() event.Type
//	DeviceId() uuid.UUID
//}
//
//type typ struct {
//	Type Type `json:"event"`
//}
//
//// Parse parses a single event object or an array of event objects
//// into a slice of Type. The slice will contain at least one element if
//// an error is not returned.
//func Parse(b []byte) ([]Type, error) {
//	b2 := bytes.TrimLeft(b, " \t\r\n")
//
//	var events []Type
//
//	switch {
//	case len(b2) > 0 && b2[0] == '{':
//		// parse event
//		e, err := parseEvent(b)
//		if err != nil {
//			return nil, err
//		}
//
//		events = append(events, e)
//	case len(b2) > 0 && b2[0] == '[':
//		// parse an array of events
//		evts, err := parseEventArray(b)
//		if err != nil {
//			return nil, err
//		}
//
//		events = evts
//	default:
//		return nil, errors.New("invalid message")
//	}
//
//	return events, nil
//}
//
//func parseEvent(b []byte) (Type, error) {
//	// parse once to get the event type
//	var et typ
//	err := json.Unmarshal(b, &et)
//	if err != nil {
//		return nil, err
//	}
//
//	e, err := FromJSON(et.Type, b)
//	if err != nil {
//		return nil, err
//	}
//
//	return e, nil
//}
//
//func parseEventArray(b []byte) ([]Type, error) {
//	var ets []typ
//	err := json.Unmarshal(b, &ets)
//	if err != nil {
//		return nil, err
//	}
//
//	events := make([]Type, len(ets))
//
//	// create new decoder to parse the actual events based on the types
//	dec := json.NewDecoder(bytes.NewReader(b))
//
//	// read the square bracket of the JSON array again
//	_, _ = dec.Token()
//
//	// for each event, decode it based on the type we parsed earlier
//	for i, typ := range ets {
//		var rm json.RawMessage
//
//		err = dec.Decode(&rm)
//		if err != nil {
//			return nil, err
//		}
//
//		e, err := FromJSON(typ.Type, rm)
//		if err != nil {
//			return nil, err
//		}
//
//		events[i] = e
//	}
//
//	return events, nil
//}
//
//func FromJSON(typ Type, b []byte) (Type, error) {
//	switch typ {
//	//case SetIdle:
//	//	var e SetIdleEvent
//	//	err := json.Unmarshal(b, &e)
//	//	return e, err
//	//case SetActive:
//	//	var e SetSinkActiveEvent
//	//	err := json.Unmarshal(b, &e)
//	//	return e, err
//	default:
//		return nil, errors.New("invalid type")
//	}
//}
