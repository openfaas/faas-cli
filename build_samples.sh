#!/bin/sh

./faas-cli -action build -yaml ./samples.yml

docker images |head -n 4
