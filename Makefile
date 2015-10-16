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
	go get -u golang.org/x/net/websocket

build: *.go
	docker build -f Dockerfile-broker -t thraxil/windsock-broker .
	docker build -f Dockerfile-windsock -t thraxil/windsock .

push: build
	docker push thraxil/windsock
	docker push thraxil/windsock-broker
