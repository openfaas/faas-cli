FROM node:6.11.2-alpine

# Alternatively use ADD https:// (which will not be cached by Docker builder)
RUN apk --no-cache add curl \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/alexellis/faas/releases/download/0.6.1/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

WORKDIR /root/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

# Wrapper/boot-strapper
COPY package.json       .
RUN npm i

# Function
COPY index.js           .
RUN mkdir -p ./root/function

COPY function/*.json    ./function/
WORKDIR /root/function
RUN npm i || :
WORKDIR /root/
COPY function           function
WORKDIR /root/function

WORKDIR /root/

ENV cgi_headers="true"

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]