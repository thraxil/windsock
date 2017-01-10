compiling (on ubuntu):

Obviously, you'll need the basic Go compiler installed and your
`GOPATH` environment variable set up. Then you just need the ZMQ
libraries available.

     $ sudo aptitude install libzmq-dev
     $ make install_deps
     $ make

running the broker:

     $ ./broker -config=broker.json

running windsock:

     $ ./windsock -config=windsock.json

The included `.json` should get you started for configuring
them. `windsock_ssl.json` is an example of SSL/WSS config. If the
'Certificate' and 'Key' fields are both present, windsock will use
SSL/WSS. Otherwise, it will be plain WS.

Obviously, you will want to make sure your configs agree on the Req +
Pub/Sub ports, and you will want to make a new "Secret" for windsock
and your client application to share. If you're using SSL, point the
Certificate and Key fields at your `.pem` and `.key` files.

These are both available on the public Docker Hub:
[thraxil/windsock](https://hub.docker.com/r/thraxil/windsock/)
and [thraxil/windsock-broker](https://hub.docker.com/r/thraxil/windsock-broker/)

The 'generate_cert.go` program is included to help you make a
self-signed certificate (for testing). It's just copied from the Go
standard library
(http://golang.org/src/pkg/crypto/tls/generate_cert.go). Compile it
with `go build generate_cert.go` and run `./generate_cert`. Don't use
self-signed certificates in production though.


The `client.py` and `sub.py` programs are included as examples of how
you might interact with windsock and the broker. `sub.py` just
subscribes to the broker and prints every message that comes across to
STDOUT (handy for debugging). `client.py` lets you type into STDIN and
sends whatever you type to the broker.


Known issues:

If you're using go version 1.4.2 with windsock, you might get this
"invalid heap pointer" error: https://github.com/pebbe/zmq2/issues/3

There are a few workarounds for this:
   * Use a different Go version.
   * **Or**, use libzmq version 4, with https://github.com/pebbe/zmq4.
     In order to do this, you need to install the `libzmq3-dev`
     (sic) package in ubuntu, and switch `pebbe/zmq2` to `pebbe/zmq4`
     in the source files.
