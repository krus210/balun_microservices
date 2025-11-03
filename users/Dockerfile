# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf make

WORKDIR /app

# Copy shared libraries
COPY lib/postgres/ lib/postgres/
COPY lib/secrets/ lib/secrets/

# Copy users service
COPY users/ users/

# Download dependencies and build
# GOWORK=off disables go.work to use go.mod replace directives
WORKDIR /app/users
ENV GOWORK=off
RUN go mod download
RUN go build -o users ./cmd/

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/users/users .

EXPOSE 8082

CMD ["./users"]
