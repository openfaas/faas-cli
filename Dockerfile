# Build stage
FROM golang:1.13 as builder

ENV GO111MODULE=off
ENV CGO_ENABLED=0

WORKDIR /usr/bin/
RUN curl -sLSf https://raw.githubusercontent.com/teamserverless/license-check/master/get.sh | sh

WORKDIR /go/src/github.com/openfaas/faas-cli
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

# ldflags "-s -w" strips binary
# ldflags -X injects commit version into binary
RUN /usr/bin/license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Author(s)"

RUN go test $(go list ./... | grep -v /vendor/ | grep -v /template/|grep -v /build/|grep -v /sample/) -cover

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
    && GIT_COMMIT=$(git rev-list -1 HEAD) \
    && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
    -X github.com/openfaas/faas-cli/version.GitCommit=${GIT_COMMIT} \
    -X github.com/openfaas/faas-cli/version.Version=${VERSION} \
    -X github.com/openfaas/faas-cli/commands.Platform=x86_64" \
    -a -installsuffix cgo -o faas-cli

# Release stage
FROM alpine:3.11 as release

RUN apk --no-cache add ca-certificates git

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk add --no-cache ca-certificates

WORKDIR /home/app

COPY --from=builder /go/src/github.com/openfaas/faas-cli/faas-cli               /usr/bin/
RUN chown -R app:app ./

USER app

ENV PATH=$PATH:/usr/bin/

CMD ["faas-cli"]