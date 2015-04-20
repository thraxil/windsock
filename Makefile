all: broker windsock

broker: broker.go
	go build broker.go

windsock: windsock.go
	go build windsock.go

clean:
	rm -f broker
	rm -f windsock

install_deps:
	go get -u github.com/pebbe/zmq2
	go get -u code.google.com/p/go.net/websocket
