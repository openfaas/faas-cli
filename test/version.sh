#!/bin/bash

FAAS_BINARY=$1
GIT_COMMIT=$(git rev-list -1 HEAD)
VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///')
VERSION=${VERSION:=dev}

echo "TEST: Version command prints GitCommit"
${FAAS_BINARY} version | tee /dev/stderr | grep "Commit: ${GIT_COMMIT}"
${FAAS_BINARY} version | tee /dev/stderr | grep "Version: ${VERSION}"
if [ $? -eq 0 ]; then
    echo OK
else
    echo FAIL
    exit 1
fi
