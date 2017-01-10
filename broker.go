package main

// just listens on a REP socket and forwards
// messages to a PUB socket

import (
	"encoding/json"
	"expvar"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

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
	LogLevel   string `envconfig:"LOG_LEVEL"`
	ExpVarPort string `envconfig:"EXPVAR_PORT" default:"8081"`
}

var startTime = time.Now().UTC()
var numMessages = expvar.NewInt("NumMessages")

func uptime() interface{} {
	uptime := time.Since(startTime)
	return int64(uptime)
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

	expvar.Publish("Uptime", expvar.Func(uptime))
	go http.ListenAndServe(":"+c.ExpVarPort, nil)
	log.Info("expvar available on :" + c.ExpVarPort + "/debug/vars")

	var e envelope
	for {
		msg, _ := repsocket.Recv(0)
		json.Unmarshal([]byte(msg), &e)
		pubsocket.Send(e.Address, zmq.SNDMORE)
		pubsocket.Send(e.Content, 0)
		repsocket.Send("published", 0)
		log.Debug("published message")
		numMessages.Add(1)
	}
}
