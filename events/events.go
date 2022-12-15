package events

type AdbEvent struct {
	Event EventType
	Item  interface{}
}

type EventType string

const (
	Connected  EventType = "Connected"
	Disconnect EventType = "Disconnected"
)
