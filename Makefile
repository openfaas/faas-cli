.PHONY: build

build:
	./build.sh

build_redist:
	./build_redist.sh

test-unit:
	OPEN_FAAS_TELEMETRY=0 go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover
