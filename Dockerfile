ARG DOCKERHUB_MIROR="docker.io"

FROM ${DOCKERHUB_MIROR}/library/golang:1.25.9-alpine AS builder

WORKDIR /workspace/packages/cli

RUN apk --no-cache add ca-certificates git

COPY ./packages/cli/go.mod ./packages/cli/go.sum ./
RUN go mod download

COPY ./packages/cli/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/escape-cli ./cmd/

FROM ${DOCKERHUB_MIROR}/library/alpine:3.23.3

RUN apk --no-cache add ca-certificates curl && \
    addgroup -g 1000 escape && \
    adduser -D -u 1000 -G escape escape

COPY --from=builder /build/escape-cli /usr/local/bin/escape-cli

USER 1000
WORKDIR /home/escape

ENTRYPOINT ["escape-cli"]
