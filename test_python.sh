#!/bin/bash

go build && ./faas-cli -action=build -image=alexellis2/hello-py -name=hello-py -handler=./sample/py -lang=python
