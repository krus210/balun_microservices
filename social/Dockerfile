FROM golang:1.25-alpine

RUN apk add --no-cache protobuf make

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o social ./cmd/

EXPOSE 8083

CMD ["./social"]
