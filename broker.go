package main

// just listens on a REP socket and forwards
// messages to a PUB socket

import (
	"encoding/json"
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"io/ioutil"
)

type envelope struct {
	Address string `json:"address"`
	Content string `json:"content"`
}

type ConfigData struct {
	RepSocket string
	PubSocket string
}

func main() {
	var configfile string
	flag.StringVar(&configfile, "config", "./broker.json", "Broker JSON config file")
	flag.Parse()

	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		fmt.Println("could not read config file")
	}

	f := ConfigData{}
	err = json.Unmarshal(file, &f)

	context, _ := zmq.NewContext()
	pubsocket, _ := context.NewSocket(zmq.PUB)
	repsocket, _ := context.NewSocket(zmq.REP)
	defer context.Close()
	defer pubsocket.Close()
	defer repsocket.Close()
	pubsocket.Bind(f.PubSocket)
	repsocket.Bind(f.RepSocket)

	var e envelope
	for {
		msg, _ := repsocket.Recv(0)
		json.Unmarshal([]byte(msg), &e)
		pubsocket.SendMultipart([][]byte{[]byte(e.Address), []byte(e.Content)}, 0)
		repsocket.Send([]byte("published"), 0)
	}
}
