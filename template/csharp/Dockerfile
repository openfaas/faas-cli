FROM microsoft/dotnet:2.0-sdk as builder

ENV DOTNET_CLI_TELEMETRY_OPTOUT 1

WORKDIR /root/
WORKDIR /root/src
COPY .  .
RUN dotnet restore ./root.csproj
RUN dotnet publish -c release -o published

FROM microsoft/dotnet:2.0-runtime

#ADD https://github.com/openfaas/faas/releases/download/0.6.1/fwatchdog /usr/bin
RUN apt-get update -qy \
    && apt-get install -qy curl ca-certificates --no-install-recommends \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.1/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apt-get -qy remove curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /root/src/published .

ENV fprocess="dotnet ./root.dll"
EXPOSE 8080
CMD ["fwatchdog"]

