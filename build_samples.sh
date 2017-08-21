#!/bin/sh

./faas-cli build --yaml ./samples.yml # --squash=true

docker images |head -n 4
