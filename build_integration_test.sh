#!/bin/bash

cli="./faas-cli"
template="python3"

get_package() {
    uname=$(uname)

    echo "Getting faas-cli package for $uname..."

    case $uname in
    "Darwin")
    cli="./faas-cli-darwin"
    ;;
    "Linux")
        arch=$(uname -m)
        echo $arch
        case $arch in
        "armv6l" | "armv7l")
        cli="./faas-cli-armhf"
        template="python3-armhf"
        ;;
        esac
    ;;
    esac

    echo "Using package $cli"
    echo "Using template $template"
}

build_faas_function() {

    function_name=$1

    eval $cli new $function_name --lang $template

cat << EOF > $function_name/handler.py
def handle(req):

    return "Function output from integration testing: Hello World!"
EOF

    eval $cli build -f $function_name.yml
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

get_package
build_faas_function $FUNCTION
run_and_test $FUNCTION $PORT $FUNCTION_UP_TIMEOUT



