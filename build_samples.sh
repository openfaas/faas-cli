#!/bin/sh

# docker rmi alexellis2/faas-get_captains
# docker rmi alexellis2/faas-urlping 
# docker rmi alexellis2/faas-node_info

./faas-cli -action build -yaml ./samples.yml

docker images |head -n 4
