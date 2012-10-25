import sys
import zmq
from simplejson import loads

PUB_KEY = "test"
PUB_SOCKET = "tcp://localhost:5556"

context = zmq.Context()
socket = context.socket(zmq.SUB)
socket.connect (PUB_SOCKET)
socket.setsockopt(zmq.SUBSCRIBE, "")

while True:
    [address, message] = socket.recv_multipart()
    print address, message


