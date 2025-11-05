# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf make

WORKDIR /app

# Copy auth service
COPY auth/ auth/

# Download dependencies and build
# GOWORK=off disables go.work to use go.mod replace directives
WORKDIR /app/auth
ENV GOWORK=off
RUN go mod download
RUN go build -o auth ./cmd/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/auth/auth .

EXPOSE 8082

CMD ["./auth"]
