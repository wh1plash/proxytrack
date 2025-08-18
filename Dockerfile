# Multi-stage build для оптимизации размера образа
FROM golang:1.24 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN apt-get update && apt-get upgrade -y && go mod download
RUN go mod download
RUN go mod verify
COPY . .
RUN go mod tidy
RUN make build

# Финальный образ
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/bin/app ./bin/app
EXPOSE 8084
CMD ["/app/bin/app"]