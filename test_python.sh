#!/bin/bash

docker service rm hello-py
go build && ./faas-cli -action=build -image=alexellis2/hello-py -name=hello-py -handler=./sample/py -lang=python && \
./faas-cli -action=deploy -image=alexellis2/hello-py -name=hello-py -lang=python
sleep 3
curl -d "hi" http://localhost:8080/function/hello-py

