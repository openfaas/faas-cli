FROM microsoft/dotnet:2.0-sdk as builder

ENV DOTNET_CLI_TELEMETRY_OPTOUT 1

# Optimize for Docker builder caching by adding projects first.

RUN mkdir -p /home/app/src/function
WORKDIR /home/app/src/function
COPY ./function/Function.csproj  .

WORKDIR /home/app/src
COPY ./root.csproj  .

RUN dotnet restore ./root.csproj

COPY .  .

RUN dotnet publish -c release -o published

FROM microsoft/dotnet:2.0-runtime
RUN addgroup -S app && adduser -S -g app app

RUN apt-get update -qy \
    && apt-get install -qy curl ca-certificates --no-install-recommends \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.12/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apt-get -qy remove curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /home/app/
COPY --from=builder /home/app/src/published .

# chmod for tmp is for a buildkit issue (@alexellis)
RUN chown app:app -R /home/app \
    && chmod 777 /tmp

USER app

ENV fprocess="dotnet ./root.dll"
EXPOSE 8080
CMD ["fwatchdog"]

