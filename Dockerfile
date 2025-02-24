FROM alpine:3.20
COPY ./escape-cli /usr/local/bin/escape-cli

RUN adduser -D escape
USER escape
ENTRYPOINT ["/usr/local/bin/escape-cli"]
