# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf make

WORKDIR /app

# Copy shared libraries
COPY lib/postgres/ lib/postgres/
COPY lib/secrets/ lib/secrets/
COPY lib/config/ lib/config/
COPY lib/app/ lib/app/
COPY lib/grpc/ lib/grpc/

# Copy gateway service
COPY gateway/ gateway/

# Download dependencies and build
# GOWORK=off disables go.work to use go.mod replace directives
WORKDIR /app/gateway
ENV GOWORK=off
RUN go mod download
RUN go build -o gateway ./cmd/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/gateway/gateway .

# Copy config file
COPY gateway/config.yaml .

EXPOSE 8080 8085

CMD ["./gateway"]
