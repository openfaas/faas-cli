FROM golang:1.8.3

WORKDIR /go/src/github.com/alexellis/faas-cli
COPY . .
RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-X main.GitCommit=${GIT_COMMIT}" -a -installsuffix cgo -o faas-cli .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 /go/src/github.com/alexellis/faas-cli/faas-cli    .

CMD ["./faas-cli"]

