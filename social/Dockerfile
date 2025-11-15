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
COPY lib/logger/ lib/logger/
COPY lib/tracer/ lib/tracer/
COPY lib/metrics/ lib/metrics/
COPY lib/admin/ lib/admin/

# Copy social service
COPY social/ social/

# Download dependencies and build
# GOWORK=off disables go.work to use go.mod replace directives
WORKDIR /app/social
ENV GOWORK=off
RUN go mod download
RUN go build -o social ./cmd/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/social/social .

# Copy config file
COPY social/config.yaml .

EXPOSE 8082

CMD ["./social"]
