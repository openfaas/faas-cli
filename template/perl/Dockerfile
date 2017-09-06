FROM perl:5.26


RUN apt-get update -qy \
    && apt-get install -qy curl ca-certificates --no-install-recommends \
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/alexellis/faas/releases/download/0.6.4/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apt-get -qy remove curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN cpanm Carton

WORKDIR /root/

COPY index.pl .
COPY cpanfile .

RUN carton install

COPY function function

WORKDIR /root/function/
COPY function/cpanfile .
RUN carton install

WORKDIR /root/

ENV fprocess="carton exec perl index.py"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
