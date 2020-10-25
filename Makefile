GO_FILES?=$$(find . -name '*.go' |grep -v vendor)

.GIT_COMMIT=$(shell git rev-parse HEAD)
.GIT_VERSION=$(shell git describe --tags --always --dirty)
.LDFLAGS=-s -w -X main.Version=$(.GIT_VERSION) -X main.GitCommit=$(.GIT_COMMIT)
.PLATFORMS=linux/amd64,linux/arm/v6,linux/arm64
.BUILDX_OUTPUT=type=image,push=false

# docker manifest command will work with Docker CLI 18.03 or newer
# but for now it's still experimental feature so we need to enable that
export DOCKER_CLI_EXPERIMENTAL=enabled

export GOFLAGS=-mod=vendor

.PHONY: build
build:
	@docker buildx create --use --name=multiarch --node multiarch
	docker buildx build \
		--progress=plain \
		--build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} \
		--build-arg VERSION=$(.GIT_VERSION) --build-arg GIT_COMMIT=$(.GIT_COMMIT) \
		--platform $(.PLATFORMS) \
		--output "$(.BUILDX_OUTPUT)" \
		--target release \
		--tag openfaas/faas-cli:latest \
		--tag openfaas/faas-cli:$(.GIT_VERSION) .
	docker buildx build \
		--progress=plain \
		--build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} \
		--build-arg VERSION=$(.GIT_VERSION) --build-arg GIT_COMMIT=$(.GIT_COMMIT) \
		--platform $(.PLATFORMS) \
		--output "$(.BUILDX_OUTPUT)" \
		--target root \
		--tag openfaas/faas-cli:latest-root \
		--tag openfaas/faas-cli:$(.GIT_VERSION)-root .

.PHONY: build_redist
build_redist:
	./build_redist.sh


# Defining the `push` target twice let's us set the BUILDX_OUTPUT variable before build is rung
push: .BUILDX_OUTPUT=type=image,push=true
push: build

.PHONY: docker-login
docker-login:
	echo -n "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin

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

.PHONY: test-unit
test-unit:
	go test $(shell go list ./... | grep -v /vendor/ | grep -v /template/ | grep -v build) -cover

.PHONY: test-templating
PORT?=38080
FUNCTION?=templating-test-func
FUNCTION_UP_TIMEOUT?=30
.EXPORT_ALL_VARIABLES:
test-templating:
	./build_integration_test.sh

