#!/bin/sh

export eTAG="latest-dev"
echo $1
if [ $1 ] ; then
  eTAG=$1
fi

# This file extracts the contents of the openfaas/faas-cli:${eTAG} image
# to the bin folder.
# Run ./build.sh first to create the image itself with the various
# binaries within it.
docker create --name faas-cli openfaas/faas-cli:${eTAG} && \
  mkdir -p ./bin && \
  docker cp faas-cli:/home/app/faas-cli ./bin && \
  docker cp faas-cli:/home/app/faas-cli-darwin ./bin && \
  docker cp faas-cli:/home/app/faas-cli-armhf ./bin && \
  docker cp faas-cli:/home/app/faas-cli-arm64 ./bin && \
  docker cp faas-cli:/home/app/faas-cli.exe ./bin && \
  docker rm -f faas-cli
