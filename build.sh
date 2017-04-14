#!/bin/sh

docker build -t faas-cli . && \
 docker create --name faas-cli faas-cli && \
 docker cp faas-cli:/root/faas-cli . && \
 docker rm -f faas-cli

