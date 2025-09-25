FROM golang:1.25-alpine

RUN apk add --no-cache protobuf make

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o gateway cmd/main.go

EXPOSE 8080 8082

CMD ["./gateway"]

