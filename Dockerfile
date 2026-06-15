ARG GO_VERSION=1.26
FROM golang:${GO_VERSION}-trixie AS builder

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app/

RUN GOOS=linux GOARCH=amd64 go build -o qordled ./cmd/qordled/main.go

FROM ubuntu:24.04

WORKDIR /qordled/

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/qordled /qordled/
COPY --from=builder /app/public /qordled/public

EXPOSE 8080

ENTRYPOINT ["/qordled/qordled", "--port", "8080"]
