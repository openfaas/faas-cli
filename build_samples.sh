#!/bin/sh

./faas-cli -action build -handler=./sample/info -name=node-info -image=node-info -lang=node
./faas-cli -action build -handler=./sample/py -name=py-hello -image=py-hello -lang=python

docker images |grep node-info
docker images |grep py-hello
