#!/bin/sh

docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t faas-cli . -f Dockerfile.redist && \
 docker create --name faas-cli faas-cli && \
 docker cp faas-cli:/root/faas-cli . && \
 docker cp faas-cli:/root/faas-cli-darwin . && \
 docker cp faas-cli:/root/faas-cli-armhf . && \
 docker cp faas-cli:/root/faas-cli-arm64 . && \
 docker cp faas-cli:/root/faas-cli.exe . && \
 docker rm -f faas-cli
