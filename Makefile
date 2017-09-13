include makefiles/*

DOCKER_IMAGE := golang:1.8.3
DOCKER_RUN := docker run --rm -v "$(CURDIR):/go/src/app" -w /go/src/app $(DOCKER_IMAGE)
DOCKER_BUILD := docker build --build-arg http_proxy=$http_proxy --build-arg https_proxy=$https_proxy -t faas-cli .

.PHONY: build build_redist build_samples deploy_samples test test_version
