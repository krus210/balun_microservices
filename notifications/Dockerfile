# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf make

WORKDIR /app

# Copy shared libraries
COPY lib/postgres/ lib/postgres/
COPY lib/secrets/ lib/secrets/

# Copy notifications service
COPY notifications/ notifications/

# Download dependencies and build
# GOWORK=off disables go.work to use go.mod replace directives
WORKDIR /app/notifications
ENV GOWORK=off
RUN go mod download
RUN go build -o notifications ./cmd/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/notifications/notifications .

EXPOSE 8082

CMD ["./notifications"]
