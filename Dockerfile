FROM golang:1.8.3

WORKDIR /go/src/github.com/openfaas/faas-cli
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))"

RUN VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') \
 && GIT_COMMIT=$(git rev-list -1 HEAD) \
 && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w -X github.com/openfaas/faas-cli/commands.GitCommit=${GIT_COMMIT} -X github.com/openfaas/faas-cli/commands.Version=${VERSION}" -a -installsuffix cgo -o faas-cli .
RUN go test $(go list ./... | grep -v /vendor/ | grep -v /template/) -cover

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 /go/src/github.com/openfaas/faas-cli/faas-cli    .

CMD ["./faas-cli"]

