FROM node:6.11.2-alpine

RUN addgroup -S app && adduser -S -g app app

# Alternatively use ADD https:// (which will not be cached by Docker builder)
RUN apk --no-cache add curl \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

WORKDIR /root/

# Turn down the verbosity to default level.
ENV NPM_CONFIG_LOGLEVEL warn

RUN mkdir -p /home/app

# Wrapper/boot-strapper
WORKDIR /home/app
COPY package.json ./

# This ordering means the npm installation is cached for the outer function handler.
RUN npm i

# Copy outer function handler
COPY index.js ./

# COPY function node packages and install, adding this as a separate
# entry allows caching of npm install
WORKDIR /home/app/function
COPY function/*.json ./
RUN npm i || :

# COPY function files and folders
COPY function/ ./

# Set correct permissions to use non root user
WORKDIR /home/app/

# chmod for tmp is for a buildkit issue (@alexellis)
RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV cgi_headers="true"
ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]

