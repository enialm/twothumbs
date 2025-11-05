# Builder
FROM golang:1.24-bookworm AS builder

WORKDIR /twothumbs

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o ingest ./cmd/ingest
RUN go build -o interact ./cmd/interact
RUN go build -o digest ./cmd/digest

# Runtime
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y pyxplot ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /twothumbs

COPY --from=builder /twothumbs/ingest .
COPY --from=builder /twothumbs/interact .
COPY --from=builder /twothumbs/digest .
