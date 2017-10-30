.PHONY: build

build:
	./build.sh

build_redist:
	./build_redist.sh

test-unit:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover
