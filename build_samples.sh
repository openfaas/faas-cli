#!/bin/sh

./faas-cli build # --squash=true

docker images |head -n 4
