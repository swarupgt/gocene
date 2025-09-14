import requests
import json

with open('sample_input.json', 'r') as file:
    data = json.load(file)

endpoint_url = 'http://localhost:8080/idx1/add_document'
i = 0

for item in data:
    print(i)
    i+=1
    d = dict()
    d["data"] = item
    response = requests.post(endpoint_url, json=d)
    print(f"Status Code: {response.status_code}, Response: {response.text}")