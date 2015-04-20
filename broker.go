package main

// just listens on a REP socket and forwards
// messages to a PUB socket

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	zmq "github.com/pebbe/zmq2"
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
		panic("could not read config file: " + err.Error())
	}

	f := ConfigData{}
	err = json.Unmarshal(file, &f)

	pubsocket, _ := zmq.NewSocket(zmq.PUB)
	repsocket, _ := zmq.NewSocket(zmq.REP)
	defer pubsocket.Close()
	defer repsocket.Close()
	pubsocket.Bind(f.PubSocket)
	repsocket.Bind(f.RepSocket)

	var e envelope
	for {
		msg, _ := repsocket.Recv(0)
		json.Unmarshal([]byte(msg), &e)
		pubsocket.Send(e.Address, zmq.SNDMORE)
		pubsocket.Send(e.Content, 0)
		repsocket.Send("published", 0)
	}
}
