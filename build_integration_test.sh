#!/bin/bash

set -e # Exit script immediately on first error.
set -x # Print commands and their arguments as they are executed.
set -o pipefail # Print commands and their arguments as they are

cli="./bin/faas-cli"

TEMPLATE_NAME="python3-http"

get_package() {
    uname=$(uname)
    arch=$(uname -m)
    echo "Getting faas-cli package for $uname..."
    echo "Having architecture $arch..."

    case $uname in
    "Darwin")
        cli="./faas-cli-darwin"
        case $arnch in
        "arm64")
        cli="./faas-cli-darwin-arm64"
        ;;
        esac
    ;;
    "Linux")
        case $arch in
        "armv6l" | "armv7l")
        cli="./faas-cli-armhf"
        ;;
        esac
    ;;
    esac

    echo "Using package $cli"
    echo "Using template $TEMPLATE_NAME"
}

build_faas_function() {

    function_name=$1

    eval $cli new $function_name --lang $TEMPLATE_NAME

cat << EOF > $function_name/handler.py
def handle(event, context):
    return {
        "statusCode": 200,
        "body": {"message": "Hello from OpenFaaS!"},
        "headers": {
            "Content-Type": "application/json"
        }
    }
EOF

    eval $cli build -f stack.yaml
}

wait_for_function_up() {
    function_name=$1
    port=$2
    timeout=$3

    function_up=false
    for (( i=1; i<=$timeout; i++ ))
    do
        echo "Checking if 127.0.0.1:$port is up.. ${i}/$timeout"
        status_code=$(curl 127.0.0.1:$port/_/health -o /dev/null -sw '%{http_code}')
        if [ $status_code == 200 ] ; then
            echo "Function $function_name is up on 127.0.0.1:$port"
            function_up=true
            break
        else
            echo "Waiting for function $function_name to run on 127.0.0.1:$port"
            sleep 1
        fi
    done

    if [ "$function_up" != true ] ; then
        echo "Failed to reach function on 127.0.0.1:$port.. Service timeout"

        echo "Removing container..."
        docker rm -f $id

        echo "Removing function image $function_name:latest"
        docker rmi -f $function_name:latest

        exit 1
    fi
}

run_and_test() {
    function_name=$1
    port=$2
    timeout=$3

    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        echo "Port $port is already allocated.. Cannot use this port for testing.
        Exiting test..."

        echo "Removing function image $function_name:latest"
        docker rmi -f $function_name:latest
        exit 1
    fi

    id=$(docker run --env fprocess="python3 index.py" --name $function_name -p $port:8080 -d $function_name:latest)

    wait_for_function_up $function_name $port $timeout

    curl -s 127.0.0.1:$port > got.txt

cat << EOF > want.txt
Function output from integration testing: Hello World!
EOF

    if cmp got.txt want.txt ; then
        echo "SUCCESS testing function $function_name"
    else
        echo "FAILED testing function $function_name"
    fi

    echo "Removing container..."
    docker rm -f $id

    echo "Removing function image $function_name:latest"
    docker rmi -f $function_name:latest

    echo "Removing created files..."
    rm -rf got.txt want.txt $function_name*
}

get_templates() {
    echo "Getting templates..."
    eval $cli template store pull $TEMPLATE_NAME
}

get_templates

get_package
build_faas_function $FUNCTION
run_and_test $FUNCTION $PORT $FUNCTION_UP_TIMEOUT



