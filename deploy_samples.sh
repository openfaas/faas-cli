#!/bin/sh

./faas-cli deploy --yaml ./stack.yml

sleep 5

# Get sample image for resizer function.
curl -SL https://raw.githubusercontent.com/openfaas/faas/master/sample-functions/ResizeImageMagick/gordon.png > gordon.png

echo "Testing nodejs-echo"
curl -sd "This was the input string." http://127.0.0.1:8080/function/nodejs-echo
echo

echo "Testing url-ping"
curl -sd "https://shop.pimoroni.com" http://127.0.0.1:8080/function/url-ping
echo

echo "Testing shrink-image"
curl -d "" http://127.0.0.1:8080/function/shrink-image --data-binary @gordon.png > small_gordon.png
echo
