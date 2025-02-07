# Runtime
FROM alpine:3.14

RUN adduser -D escape
USER escape

COPY --chown=escape:escape ./escape-cli /usr/local/bin/escape-cli

ENTRYPOINT ["/bin/sh"]