package main

// just listens on a REP socket and forwards
// messages to a PUB socket

import (
	"encoding/json"
	zmq "github.com/alecthomas/gozmq"
)

var REP_SOCKET = "tcp://*:5555"
var PUB_SOCKET = "tcp://*:5556"

type envelope struct {
	Address string `json:"address"`
	Content string `json:"content"`
}

func main() {
	context, _ := zmq.NewContext()
	pubsocket, _ := context.NewSocket(zmq.PUB)
	repsocket, _ := context.NewSocket(zmq.REP)
	defer context.Close()
	defer pubsocket.Close()
	defer repsocket.Close()
	pubsocket.Bind(PUB_SOCKET)
	repsocket.Bind(REP_SOCKET)

	var e envelope
	for {
		msg, _ := repsocket.Recv(0)
		json.Unmarshal([]byte(msg), &e)
		pubsocket.SendMultipart([][]byte{[]byte(e.Address), []byte(e.Content)}, 0)
		repsocket.Send([]byte("published"), 0)
	}
}
