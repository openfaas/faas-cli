FROM arm32v6/alpine:3.6

RUN addgroup -S app && adduser -S -g app app

RUN apk add --no-cache nodejs nodejs-npm ca-certificates
RUN apk --no-cache add curl \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog-armhf > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

ENV NPM_CONFIG_LOGLEVEL warn

RUN mkdir -p /home/app

WORKDIR /home/app
COPY package.json   .

RUN npm i

COPY index.js       .

WORKDIR /home/app/function
COPY function/*.json ./
RUN npm i || :

COPY function/ ./

WORKDIR /home/app

RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV cgi_headers="true"

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
