#!/bin/bash

docker service rm hello-captains
docker rmi alexellis2/getcaptains -f
go build && \
./faas-cli -action=build -image=alexellis2/getcaptains-js -name=hello-captains -handler=./sample/getCaptains && \
./faas-cli -action=deploy -image=alexellis2/getcaptains-js -name=hello-captains

sleep 5

curl -d "Hi" http://localhost:8080/function/hello-captains
curl -d "Hi" http://localhost:8080/function/hello-captains
