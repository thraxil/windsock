import zmq
from json import dumps

context = zmq.Context()
socket = context.socket(zmq.REQ)
socket.connect ("tcp://localhost:5555")

while True:
    message = raw_input("> ")
    m = dict(message_type="message", nick="cli client", content=message)
    e = dict(address="gobot.cli",content=dumps(m))
    socket.send(dumps(e))
    socket.recv()

