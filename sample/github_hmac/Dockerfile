FROM golang:1.7.5-alpine

MAINTAINER alexellis2@gmail.com
ENTRYPOINT []
#RUN mkdir -p /go/src/github.com/openfaas/faas/sample-functions/handler
WORKDIR /go/src/github.com/openfaas/faas/sample-functions/handler
COPY . /go/src/github.com/openfaas/faas/sample-functions/handler
RUN go get -d -v
RUN go install

ADD https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV cgi_headers true
ENV fprocess "/go/bin/handler"
HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD [ "/usr/bin/fwatchdog"]
