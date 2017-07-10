#!/bin/sh

docker build -f Dockerfile.redist --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t faas-cli . && \
 docker create --name faas-cli faas-cli && \
 docker cp faas-cli:/root/faas-cli . && \
 docker cp faas-cli:/root/faas-cli-darwin . && \
 docker cp faas-cli:/root/faas-cli-armhf . && \
 docker rm -f faas-cli

