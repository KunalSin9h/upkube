# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (for go mod) and ca-certificates
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG GOVERSION=go

RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.goVersion=${GOVERSION}'" -o upkube .

# Final image
FROM alpine:latest

WORKDIR /app

# Copy binary and static assets
COPY --from=builder /app/upkube /app/upkube

EXPOSE 8080

ENV HOST=0.0.0.0
ENV PORT=8080

CMD ["/app/upkube"]