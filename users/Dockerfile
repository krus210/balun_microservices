FROM golang:1.25-alpine

RUN apk add --no-cache protobuf make

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o users ./cmd/

EXPOSE 8082

CMD ["./users"]
