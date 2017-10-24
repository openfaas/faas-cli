FROM ruby:2.4-alpine3.6

# Alternatively use ADD https:// (which will not be cached by Docker builder)
RUN apk --no-cache add curl \ 
    && echo "Pulling watchdog binary from Github." \
    && curl -sSL https://github.com/openfaas/faas/releases/download/0.6.9/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && apk del curl --no-cache

WORKDIR /root/

COPY Gemfile		.
COPY index.rb           .
COPY function           function
RUN bundle install 

WORKDIR /root/function/
RUN bundle install 

WORKDIR /root/
ENV fprocess="ruby index.rb"
HEALTHCHECK --interval=2s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]
