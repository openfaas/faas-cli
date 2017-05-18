#!/bin/sh

./faas-cli -action deploy -yaml ./samples.yml

curl -d "" http://localhost:8080/function/get_captains
echo

curl -d "This was the input string." http://localhost:8080/function/node_info
echo

curl -d "https://shop.pimoroni.com" http://localhost:8080/function/url_ping
echo
