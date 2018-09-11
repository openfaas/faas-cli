GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

.PHONY: build
build:
	./build.sh

.PHONY: build_redist
build_redist:
	./build_redist.sh

.PHONY: build_samples
build_samples:
	./build_samples.sh

.PHONY: local-fmt
local-fmt:
	gofmt -l -d $(GO_FILES)

.PHONY: local-goimports
local-goimports:
	goimports -w $(GO_FILES)

.PHONY: test-unit
test-unit:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover

ci-armhf-push:
	(docker push openfaas/faas-cli:$(TAG)-armhf)
ci-armhf-build:
	(./build.sh $(TAG)-armhf)

.PHONY: test-templating
PORT?=38080
FUNCTION?=templating-test-func
FUNCTION_UP_TIMEOUT?=30
.EXPORT_ALL_VARIABLES:
test-templating:
	./build_integration_test.sh

