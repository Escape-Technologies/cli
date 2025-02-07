# Runtime
FROM --platform=$BUILDPLATFORM alpine:3.20

RUN adduser -D escape
USER escape

COPY --chown=escape:escape ./escape-cli /usr/local/bin/escape-cli

ENTRYPOINT ["/bin/sh"]