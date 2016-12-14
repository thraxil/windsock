package main

// just listens on a REP socket and forwards
// messages to a PUB socket

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
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

type config struct {
	LogLevel string `envconfig:"LOG_LEVEL"`
}

func main() {
	log.SetLevel(log.InfoLevel)
	var configfile string
	flag.StringVar(&configfile, "config", "./broker.json", "Broker JSON config file")
	flag.Parse()

	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatal("could not read config file: " + err.Error())
	}

	f := ConfigData{}
	err = json.Unmarshal(file, &f)
	if err != nil {
		log.Fatal("could not parse config file: " + err.Error())
	}
	var c config
	err = envconfig.Process("broker", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	// defaults to INFO
	if c.LogLevel == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	}
	if c.LogLevel == "WARN" {
		log.SetLevel(log.WarnLevel)
	}
	if c.LogLevel == "ERROR" {
		log.SetLevel(log.ErrorLevel)
	}
	if c.LogLevel == "FATAL" {
		log.SetLevel(log.FatalLevel)
	}

	pubsocket, _ := zmq.NewSocket(zmq.PUB)
	repsocket, _ := zmq.NewSocket(zmq.REP)
	defer pubsocket.Close()
	defer repsocket.Close()
	pubsocket.Bind(f.PubSocket)
	repsocket.Bind(f.RepSocket)

	log.Info("listening on ZMQ sockets")

	var e envelope
	for {
		msg, _ := repsocket.Recv(0)
		json.Unmarshal([]byte(msg), &e)
		pubsocket.Send(e.Address, zmq.SNDMORE)
		pubsocket.Send(e.Content, 0)
		repsocket.Send("published", 0)
		log.Debug("published message")
	}
}
