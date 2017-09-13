USER := $(shell id -u)
IMAGE := golang:1.7
PROJECT := github.com/alexellis/faas-cli

BUILD_CONFIG := CGO_ENABLED=0 GOOS=linux go build --ldflags "-X main.GitCommit=${GIT_COMMIT}" -a -installsuffix cgo -o faas-cli
DOCKER_RUN := docker run --rm -u $(USER) -v $(PWD):/go/src/$(PROJECT) -w /go/src/$(PROJECT) $(IMAGE)

clean:
	@echo "Cleaning up after ourselves ..."
	@$(DOCKER_RUN) go clean

get:
	@echo "Fetching dependencies ..."
	@$(DOCKER_RUN) go get -d -v

build:
	@echo "Building faas-cli ..."
	@$(DOCKER_RUN) CGO_ENABLED=0 GOOS=linux go build --ldflags "-X main.GitCommit=$(GIT_COMMIT)" -a in

test:
	@echo "Running tests ..."
	@$(DOCKER_RUN) go test

.PHONY: get build
