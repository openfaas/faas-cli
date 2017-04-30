#!/bin/bash

docker service rm hello-py
docker rmi alexellis2/hello-py -f
go build && \
./faas-cli -action=build -image=alexellis2/hello-py -name=hello-py -handler=./sample/py -lang=python && \
./faas-cli -action=deploy -image=alexellis2/hello-py -name=hello-py -lang=python

sleep 5

curl -d "Hi" http://localhost:8080/function/hello-py
curl -d "http://docs.get-faas.com" http://localhost:8080/function/hello-py
curl -d "http://faaster.io" http://localhost:8080/function/hello-py # this will timeout
