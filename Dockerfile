FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
RUN go build -v ./cmd/goodmorningbot

FROM --platform=$BUILDPLATFORM alpine:latest

WORKDIR /app

COPY --from=builder /app/goodmorningbot /app/goodmorningbot

ENTRYPOINT '/app/goodmorningbot'
