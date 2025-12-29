import socket
import json
import requests

request = {
    "method": "HelloService.Hello",
    "params": ["test"],
    "id": 0
}


res = requests.post(url="http://127.0.0.1:8082/jsonRpc", json=request)
print(res.text)

