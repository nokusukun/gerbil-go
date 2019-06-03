package gerbil

import (
	"fmt"
	"time"
)

func NewEventHandler() *EventHandler {
	handler := EventHandler{}
	handler.Events = make(map[string]func(*GerbCon, []byte, Packet) error)
	return &handler
}

type EventHandler struct {
	Events map[string]func(*GerbCon, []byte, Packet) error
}

func (e *EventHandler) Event(eventName string, function func(*GerbCon, []byte, Packet) error) {
	e.Events[eventName] = function
}

type Packet struct {
	Name     string      `json:"name"`
	Payload  interface{} `json:"payload"`
	ReturnID string      `json:"retid"`
}

type EventPacket struct {
	EventName string `json:"eventName"`
	Data      string `json:"data"`
}

type ErrorPacket struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func MakeSnowflake(prefix uint64) string {
	ts := time.Now().UnixNano()
	if prefix > 999 {
		prefix = 999
	}
	if prefix < 100 {
		prefix = 100
	}
	return fmt.Sprintf("%v", (prefix*1000000000000000)+(uint64(ts)/100000)+1000000000000000000)
}
