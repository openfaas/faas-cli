.PHONY: build

build:
	./build.sh

build_redist:
	./build_redist.sh

unit_test:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/) -cover
