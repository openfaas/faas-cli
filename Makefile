GO_FILES?=$$(find . -name '*.go' |grep -v vendor)
TAG?=latest

.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_VERSION=$(shell git describe --tags 2>/dev/null || echo "$(.GIT_COMMIT)")
.GIT_UNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(.GIT_UNTRACKEDCHANGES),)
	.GIT_COMMIT := $(.GIT_COMMIT)-dirty
endif

export GOFLAGS=-mod=vendor

.PHONY: build
build:
	./build.sh

.PHONY: build_redist
build_redist:
	./extract_binaries.sh

.PHONY: build_samples
build_samples:
	./build_samples.sh

.PHONY: local-fmt
local-fmt:
	gofmt -l -d $(GO_FILES)

.PHONY: local-goimports
local-goimports:
	goimports -w $(GO_FILES)

.PHONY: local-install
local-install:
	CGO_ENABLED=0 go install --ldflags "-s -w \
	   -X github.com/openfaas/faas-cli/version.GitCommit=${.GIT_COMMIT} \
	   -X github.com/openfaas/faas-cli/version.Version=${.GIT_VERSION}" \
	   -a -installsuffix cgo

.PHONY: dist
dist:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-s -w \
	   -X github.com/openfaas/faas-cli/version.GitCommit=${.GIT_COMMIT} \
	   -X github.com/openfaas/faas-cli/version.Version=${.GIT_VERSION}" \
	   -a -installsuffix cgo -o ./bin/faas-cli

	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build --ldflags "-s -w \
	   -X github.com/openfaas/faas-cli/version.GitCommit=${.GIT_COMMIT} \
	   -X github.com/openfaas/faas-cli/version.Version=${.GIT_VERSION}" \
	   -a -installsuffix cgo -o ./bin/faas-cli-darwin


	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build --ldflags "-s -w \
	   -X github.com/openfaas/faas-cli/version.GitCommit=${.GIT_COMMIT} \
	   -X github.com/openfaas/faas-cli/version.Version=${.GIT_VERSION}" \
	   -a -installsuffix cgo -o ./bin/faas-cli.exe


.PHONY: test-unit
test-unit:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover

.PHONY: ci-armhf-push
ci-armhf-push:
	(docker push openfaas/faas-cli:$(TAG)-armhf && docker push openfaas/faas-cli:$(TAG)-root-armhf)

.PHONY: ci-armhf-build
ci-armhf-build:
	(./build.sh $(TAG)-armhf)

.PHONY: ci-arm64-push
ci-arm64-push:
	(docker push openfaas/faas-cli:$(TAG)-arm64 && docker push openfaas/faas-cli:$(TAG)-root-arm64)

.PHONY: ci-arm64-build
ci-arm64-build:
	(./build.sh $(TAG)-arm64)

.PHONY: test-templating
PORT?=38080
FUNCTION?=templating-test-func
FUNCTION_UP_TIMEOUT?=30
.EXPORT_ALL_VARIABLES:
test-templating:
	./build_integration_test.sh

