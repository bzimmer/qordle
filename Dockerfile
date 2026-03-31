ARG GO_VERSION
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o qordled ./cmd/qordled/main.go

FROM ubuntu:24.04

WORKDIR /root/

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/qordled .
COPY --from=builder /app/public ./public

EXPOSE 8091

CMD ["./qordled", "--port", "8091"]
