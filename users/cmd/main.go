package main

import (
	"log"
	"net"
)

func main() {
	server, err := InitializeApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
