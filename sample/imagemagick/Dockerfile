FROM alpine:3.10

# Alternatively use ADD https:// (which will not be cached by Docker builder)
RUN apk --no-cache add curl ca-certificates imagemagick \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.7.0/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

ENV fprocess="convert - -resize 50% fd:1"

CMD ["fwatchdog"]