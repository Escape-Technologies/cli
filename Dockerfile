# Builder
FROM golang:1.21.0-alpine AS builder

WORKDIR /usr/src

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION
ARG COMMIT

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w \
    -X github.com/Escape-Technologies/cli/pkg/cli.version=${VERSION} \
    -X github.com/Escape-Technologies/cli/pkg/cli.commit=${COMMIT}" \
    -v -o /usr/local/bin/escape-cli ./cmd

# Runtime
FROM alpine:3.14

RUN adduser -D escape
USER escape

COPY --chown=escape:escape --from=builder /usr/local/bin/escape-cli /usr/local/bin/escape-cli 

ENTRYPOINT ["/bin/sh"]