# Build stage
FROM golang:1.9.2 as builder

RUN curl -sL \
    https://github.com/alexellis/license-check/releases/download/0.2.2/license-check > \ 
    /usr/bin/license-check \
    && chmod +x /usr/bin/license-check

WORKDIR /go/src/github.com/openfaas/faas-cli
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

# ldflags "-s -w" strips binary
# ldflags -X injects commit version into binary
RUN license-check -path ./ --verbose=false "Alex Ellis" "OpenFaaS Author(s)" "OpenFaaS Project"
RUN go test $(go list ./... | grep -v /vendor/ | grep -v /template/|grep -v /build/) -cover \
 && VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
 && GIT_COMMIT=$(git rev-list -1 HEAD) \
 && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
    -X github.com/openfaas/faas-cli/version.GitCommit=${GIT_COMMIT} \
    -X github.com/openfaas/faas-cli/version.Version=${VERSION}" \
    -a -installsuffix cgo -o faas-cli

# Release stage
FROM alpine:3.7

RUN apk --no-cache add ca-certificates git

WORKDIR /root/

COPY --from=builder /go/src/github.com/openfaas/faas-cli/faas-cli               .

ENV PATH=$PATH:/root/

CMD ["faas-cli"]

