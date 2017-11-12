FROM armhf/python:2.7-alpine
# Alternatively use ADD https:// (which will not be cached by Docker builder)

RUN addgroup -S app && adduser -S -g app app

RUN apk --no-cache add curl \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog-armhf > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

RUN mkdir -p /home/app

WORKDIR /home/app
COPY requirements.txt   .

RUN pip install -r requirements.txt

COPY index.py           .


WORKDIR /home/app/function
RUN touch ./function/__init__.py
COPY function/requirements.txt	.
RUN pip install -r requirements.txt

COPY function/ ./

WORKDIR /home/app

RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV fprocess="python index.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
