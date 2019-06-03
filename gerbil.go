package gerbil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type GerbilResponse struct {
	Type       string
	ReturnID   string
	Connection *GerbCon
	Data       interface{}
}

type GerbilBroadcast struct {
	Packet  Packet
	Channel string
}

type Gerbil struct {
	Connections  map[string]*GerbCon
	EventHandler *EventHandler
	Responses    chan *GerbilResponse
	Broadcasts   chan *GerbilBroadcast
}

// New creates a new Gerbil instance
func New() *Gerbil {
	newg := Gerbil{}
	newg.Connections = make(map[string]*GerbCon)
	newg.EventHandler = NewEventHandler()
	newg.Responses = make(chan *GerbilResponse, 500)
	newg.Broadcasts = make(chan *GerbilBroadcast, 500)
	go newg.ResponseHandler()
	go newg.BroadcastHandler()
	return &newg
}

func (g *Gerbil) ResponseHandler() {
	for resp := range g.Responses {
		resp.Connection.Connection.WriteJSON(
			Packet{
				Name:     resp.Type,
				Payload:  resp.Data,
				ReturnID: resp.ReturnID,
			},
		)
	}
}

func (g *Gerbil) BroadcastHandler() {
	for resp := range g.Broadcasts {
		fmt.Println("New broadcast recieved")
		for _, gerb := range g.Connections {
			broadcastable := stringInSlice(resp.Channel, gerb.Channels)
			if broadcastable || resp.Channel == "all" {
				gerb.Connection.WriteJSON(resp.Packet)
			}
		}
	}
}

// Event registers and event to the gerbil instance
func (g *Gerbil) Event(eventName string, function func(*GerbCon, []byte, Packet) error) {
	g.EventHandler.Event(eventName, function)
}

// Serve opens a websocket connection over to the specified address/gerbil
func (g *Gerbil) Serve(address string) {
	http.HandleFunc("/gerbil", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Failed to establish connection:", err)
			return
		}
		gerb := WrapConnection(conn, g)
		g.Connections[gerb.ID] = gerb
		fmt.Println("Established Connection:", gerb.ID)

		defer func() {
			delete(g.Connections, gerb.ID)
			gerb.Close()
		}()

		for {
			_, msg, err := gerb.Connection.ReadMessage()
			if err != nil {
				gerb.Connection.Close()
				return
			}

			// Decode recieved message as packet
			packet := Packet{}
			if err := json.Unmarshal(msg, &packet); err != nil {
				gerb.Connection.WriteJSON(Packet{
					Name: "error",
					Payload: ErrorPacket{
						Message: "Failed to unmarshal Packet",
						Data:    fmt.Sprintf("json.Unmarshal: %v", err),
					},
				})
			}

			// Process event packets
			if packet.Name == "event" {
				fmt.Println("Event recieved:", packet)
				payload := EventPacket{}
				jbytes, _ := json.Marshal(packet.Payload)
				jdata, _ := json.Marshal(packet.Payload.(map[string]interface{})["data"])

				json.Unmarshal(jbytes, &payload)

				event, exists := g.EventHandler.Events[payload.EventName]
				if !exists {
					g.Responses <- &GerbilResponse{
						ReturnID:   packet.ReturnID,
						Connection: gerb,
						Data: Packet{
							Name: "error",
							Payload: ErrorPacket{
								Message: "No event",
								Data:    fmt.Sprintf("event %v does not exist", payload.EventName),
							},
							ReturnID: packet.ReturnID,
						},
					}
				} else {
					event(gerb, []byte(jdata), packet)
				}
			}

		}
	})

	fmt.Printf("Serving on %v\n", address)
	http.ListenAndServe(address, nil)
}
