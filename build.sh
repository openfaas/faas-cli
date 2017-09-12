#!/bin/sh

# gofmt
test 0 -ne `gofmt -l . | wc -l` && echo "gofmt needed for:\n" && gofmt -l . && exit 1

# build
docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t faas-cli . && \
 docker create --name faas-cli faas-cli && \
 docker cp faas-cli:/root/faas-cli . && \
 docker rm -f faas-cli
