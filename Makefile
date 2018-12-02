.PHONY: build build_redist build_samples local-fmt local-goimports test-unit ci-armhf-push ci-armhf-build ci-arm64-push ci-arm64-build test-templating
GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

build:
	./build.sh

build_redist:
	./build_redist.sh

build_samples:
	./build_samples.sh

local-fmt:
	gofmt -l -d $(GO_FILES)

local-goimports:
	goimports -w $(GO_FILES)

test-unit:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover

ci-armhf-push:
	(docker push openfaas/faas-cli:$(TAG)-armhf)

ci-armhf-build:
	(./build.sh $(TAG)-armhf)

ci-arm64-push:
	(docker push openfaas/faas-cli:$(TAG)-arm64)

ci-arm64-build:
	(./build.sh $(TAG)-arm64)

PORT?=38080
FUNCTION?=templating-test-func
FUNCTION_UP_TIMEOUT?=30
.EXPORT_ALL_VARIABLES:
test-templating:
	./build_integration_test.sh

