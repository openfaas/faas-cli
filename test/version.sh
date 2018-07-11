#!/bin/bash

FAAS_BINARY=$1
GIT_COMMIT=$(git rev-list -1 HEAD)

echo "TEST: Version command prints GitCommit"
${FAAS_BINARY} version | tee /dev/stderr | grep "commit:  ${GIT_COMMIT}"
if [ $? -eq 0 ]; then
    echo OK
else
    echo FAIL
    exit 1
fi
