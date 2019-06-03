package gerbil

import "github.com/gorilla/websocket"

type GerbCon struct {
	Connection *websocket.Conn
	Gerbil     *Gerbil
	ID         string
	Channels   []string
}

func WrapConnection(conn *websocket.Conn, g *Gerbil) *GerbCon {
	return &GerbCon{
		Connection: conn,
		ID:         MakeSnowflake(100),
		Gerbil:     g,
	}
}

func (gc *GerbCon) Close() {
	gc.Connection.Close()
}

func (gc *GerbCon) Reply(packet Packet, payload interface{}) {
	gc.Gerbil.Responses <- &GerbilResponse{
		Type:       "reply",
		ReturnID:   packet.ReturnID,
		Connection: gc,
		Data:       payload,
	}
}

func (gc *GerbCon) Emit(payload interface{}) {
	gc.Gerbil.Responses <- &GerbilResponse{
		Type:       "event",
		Connection: gc,
		Data:       payload,
	}
}

func (gc *GerbCon) Broadcast(channel string, data interface{}) {
	gc.Gerbil.Broadcasts <- &GerbilBroadcast{
		Channel: channel,
		Packet: Packet{
			Name:    "broadcast",
			Payload: data,
		},
	}
}

func (gc *GerbCon) JoinRoom(channel string) {
	if !stringInSlice(channel, gc.Channels) {
		gc.Channels = append(gc.Channels, channel)
	}
}

func (gc *GerbCon) LeaveRoom(channel string) {
	if stringInSlice(channel, gc.Channels) {
		gc.Channels = removeSliceStr(gc.Channels, idxInSlice(channel, gc.Channels))
	}
}
