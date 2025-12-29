import socket
import json

request = {
    "method": "HelloService.Hello",
    "params": ["CHENG LIANG"],
    "id": 0
}
client = socket.create_connection(("localhost", 1234))

client.sendall(json.dumps(request).encode())

res = client.recv(1096)
print(json.loads(res.decode()))