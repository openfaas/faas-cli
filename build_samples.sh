#!/bin/sh

./faas-cli -action build -yaml ./samples.yml # -squash=true

docker images |head -n 4
