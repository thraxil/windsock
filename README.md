compiling (on ubuntu):

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

The 'generate_cert.go` program is included to help you make a
self-signed certificate (for testing). It's just copied from the Go
standard library
(http://golang.org/src/pkg/crypto/tls/generate_cert.go). Compile it
with `go build generate_cert.go` and run `./generate_cert`. Don't use
self-signed certificates in production though.