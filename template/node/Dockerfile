FROM node:6.9.1-alpine

ADD https://github.com/alexellis/faas/releases/download/0.5.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /root/

COPY package.json   .

RUN npm i
COPY index.js       .
COPY function       function
WORKDIR /root/function
RUN npm i || :
WORKDIR /root/

ENV fprocess="node index.js"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
