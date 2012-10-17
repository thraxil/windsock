package main

import (
	"code.google.com/p/go.net/websocket"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var PUB_KEY = "gobot"
// obviously, this should not be hard-coded in real life:
var SECRET = "6f1d916c-7761-4874-8d5b-8f8f93d20bf2"

type room struct {
	Users     map[*OnlineUser]bool
	Broadcast chan OutgoingMessage
	Incoming  chan Message
}

// what gets sent out to the browser
type OutgoingMessage struct {
	Time    time.Time
	Nick    string
	Content string
}

// what comes in/out of zmq 
type Message struct {
	Type    string `json:"message_type"`
	Content string `json:"content"`
	Nick    string `json:"nick"`
}

var runningRoom *room = &room{}

// listen for messages on a channel
// and fan them out them to every user in the room
func (r *room) run() {
	for b := range r.Broadcast {
		for u := range r.Users {
			u.Send <- b
		}
	}
}

func (r *room) SendLine(line OutgoingMessage) {
	r.Broadcast <- line
}

func InitRoom() {
	runningRoom = &room{
		Users:     make(map[*OnlineUser]bool),
		Broadcast: make(chan OutgoingMessage),
		Incoming:  make(chan Message),
	}
	go runningRoom.run()
}

type OnlineUser struct {
	Connection *websocket.Conn
	Nick       string
	Send       chan OutgoingMessage
}

// loop indefinitely, taking messages on a channel
// and sending them out to the user's websocket
func (this *OnlineUser) PushToClient() {
	for b := range this.Send {
		err := websocket.JSON.Send(this.Connection, b)
		if err != nil {
			break
		}
	}
}

// loop indefinitely listening for incoming
// messages from a user's websocket
func (this *OnlineUser) PullFromClient() {
	for {
		var content string
		err := websocket.Message.Receive(this.Connection, &content)

		if err != nil {
			return
		}
		runningRoom.Incoming <- Message{"msg", content, this.Nick}
		// need to echo back to ourself
		msg := OutgoingMessage{time.Now(), this.Nick, content}
		runningRoom.SendLine(msg)
	}
}

func validateToken(token string, current_time time.Time) (string, error) {
	// token will look something like this:
	// anp8:1344361884:667494:127.0.0.1:306233f64522f1f970fc62fb3cf2d7320c899851
	parts := strings.Split(token, ":")
	if len(parts) != 5 {
		return "", errors.New("invalid token")
	}
	// their UNI
	uni := parts[0]
	// UNIX timestamp
	now, err := strconv.Atoi(parts[1])
	if err != nil {
		return uni, errors.New("invalid timestamp in token")
	}
	// a random salt 
	salt := parts[2]
	ip_address := parts[3]
	// the hmac of those parts with our shared secret
	hmc := parts[4]

	// make sure we're within a 60 second window
	token_time := time.Unix(int64(now), 0)
	if current_time.Sub(token_time) > time.Duration(60*time.Second) {
		return "", errors.New("stale token")
	}
	// TODO: check that their ip address matches

	// check that the HMAC matches
	h := hmac.New(
		sha1.New,
		[]byte(SECRET))
	h.Write([]byte(fmt.Sprintf("%s:%d:%s:%s", uni, now, salt, ip_address)))
	sum := fmt.Sprintf("%x", h.Sum(nil))
	if sum != hmc {
		return "", errors.New("token HMAC doesn't match")
	}
	return uni, nil
}

func BuildConnection(ws *websocket.Conn) {
	token := ws.Request().URL.Query().Get("token")
	uni, err := validateToken(token, time.Now())
	if err != nil {
		fmt.Println("validation error: " + err.Error())
		return
	}

	onlineUser := &OnlineUser{
		Connection: ws,
		Nick:       uni,
		Send:       make(chan OutgoingMessage, 256),
	}
	runningRoom.Users[onlineUser] = true
	go onlineUser.PushToClient()
	fmt.Printf("%s joined\n", uni)
	runningRoom.Incoming <- Message{"status", uni, "joined as web user"}
	onlineUser.PullFromClient()
	fmt.Printf("%s disconnected\n", uni)
	runningRoom.Incoming <- Message{"status", uni, "web user disconnected"}
	delete(runningRoom.Users, onlineUser)
}

func receiveZmqMessage(subsocket zmq.Socket, m *Message) error {
		// using zmq multi-part messages which will arrive
		// in pairs. the first of which we don't care about so we discard.
		_, _ = subsocket.Recv(0)
    content, _ := subsocket.Recv(0)
		return json.Unmarshal([]byte(content), m)
}


// listen on a zmq SUB socket
// and shovel messages from it out to the websockets
func zmqToWebsocket(subsocket zmq.Socket) {
	var m Message
	for {
		err := receiveZmqMessage(subsocket, &m)
		if err != nil {
			// just ignore any invalid messages
			continue
		}

		if m.Type != "message" {
			fmt.Printf("can only handle messages right now")
			continue
		}
		// turn it into a proper outgoing message and send it
		msg := OutgoingMessage{time.Now(), m.Nick, m.Content}
		runningRoom.SendLine(msg)
	}
}

// send a message to the zmq PUB socket
func sendMessage(pubsocket zmq.Socket, m Message) {
	b, _ := json.Marshal(m)
	pubsocket.SendMultipart([][]byte{[]byte(PUB_KEY),b},0)
}

// take messages from the Incoming channel
// and just shovel them out to the zmq PUB socket
func websocketToZmq(pubsocket zmq.Socket) {
	for msg := range runningRoom.Incoming {
		var mtype = "message"
		if msg.Type == "notice" {
			mtype = msg.Type
		}
		sendMessage(pubsocket, Message{mtype, msg.Nick, msg.Content})
	}
}

func main() {
	context, _ := zmq.NewContext()
	pubsocket, _ := context.NewSocket(zmq.PUB)
	subsocket, _ := context.NewSocket(zmq.SUB)
	defer context.Close()
	defer pubsocket.Close()
	defer subsocket.Close()
	pubsocket.Bind("tcp://*:5557")
	subsocket.SetSockOptString(zmq.SUBSCRIBE, PUB_KEY)
	subsocket.Connect("tcp://localhost:5556")

	InitRoom()

	// two goroutines to move messages in each direction
	go websocketToZmq(pubsocket)
	go zmqToWebsocket(subsocket)

	http.Handle("/socket/", websocket.Handler(BuildConnection))
	err := http.ListenAndServe(":5050", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
