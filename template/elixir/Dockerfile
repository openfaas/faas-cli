FROM elixir:1.5-alpine

RUN apk --no-cache add curl \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.6b/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

ENV cgi_headers="true"
ENV fprocess="/bin/openfaas_function"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

# Deps / Compilation
ONBUILD COPY . /root/function
ONBUILD WORKDIR /root/function

ONBUILD RUN mix do local.hex --force, local.rebar --force, deps.get, deps.compile --all
ONBUILD RUN mix escript.build
ONBUILD RUN mv openfaas_function /bin/openfaas_function
