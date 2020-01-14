#!/bin/sh

export eTAG="latest-dev"
echo $1
if [ $1 ] ; then
  eTAG=$1
fi

docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t openfaas/faas-cli:$eTAG . -f Dockerfile.redist && \
 docker create --name faas-cli openfaas/faas-cli:$eTAG && \
 docker cp faas-cli:/home/app/faas-cli . && \
 docker cp faas-cli:/home/app/faas-cli-darwin . && \
 docker cp faas-cli:/home/app/faas-cli-armhf . && \
 docker cp faas-cli:/home/app/faas-cli-arm64 . && \
 docker cp faas-cli:/home/app/faas-cli.exe . && \
 docker rm -f faas-cli
