FROM alpine:3.6

# Alternatively use ADD https:// (which will not be cached by Docker builder)
RUN apk --no-cache add curl ca-certificates imagemagick \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/alexellis/faas/releases/download/0.5.8-alpha/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog

ENV fprocess="convert - -resize 50% fd:1"

CMD ["fwatchdog"]