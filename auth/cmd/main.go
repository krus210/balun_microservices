package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "auth/pkg/api"
)

type User struct {
	id           int64
	email        string
	password     string
	accessToken  string
	refreshToken string
}

type Server struct {
	pb.AuthServiceServer

	// Бизнес логика/зовисимоти
	mx    sync.RWMutex
	users map[string]*User
}

func NewServer() *Server {
	return &Server{users: make(map[string]*User)}
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	if err := s.validateCredentials(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, err
	}

	var err error
	var userId int64
	s.mx.Lock()
	if user, ok := s.users[req.Email]; ok && user != nil {
		err = status.New(codes.AlreadyExists, "user already exists").Err()
	}
	if err == nil {
		userId = rand.Int64()
		s.users[req.Email] = &User{
			id:       userId,
			email:    req.Email,
			password: req.Password,
		}
	}
	defer s.mx.Unlock()

	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{UserId: userId}, nil
}

func (s *Server) validateCredentials(email string, password string) error {
	var violations []*errdetails.BadRequest_FieldViolation
	if len(email) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "email",
			Description: "empty",
		})
	}
	if len(password) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "password",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		rpcErr := status.New(codes.InvalidArgument, "почта или пароль пустые")

		detailedError, err := rpcErr.WithDetails(&errdetails.BadRequest{
			FieldViolations: violations,
		})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return detailedError.Err()
	}

	return nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	if err := s.validateCredentials(req.GetEmail(), req.GetPassword()); err != nil {
		return nil, err
	}

	var err error

	s.mx.Lock()
	user, ok := s.users[req.Email]
	switch {
	case !ok || user == nil:
		err = status.New(codes.Unauthenticated, "user not exists").Err()
	case user.email != req.Email || user.password != req.Password:
		err = status.New(codes.Unauthenticated, "email or password are not right").Err()
	}
	if err == nil {
		s.users[req.Email].accessToken = uuid.New().String()
		s.users[req.Email].refreshToken = uuid.New().String()
	}
	defer s.mx.Unlock()

	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  s.users[req.Email].accessToken,
		RefreshToken: s.users[req.Email].refreshToken,
		UserId:       s.users[req.Email].id,
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	var err error
	var resultUser *User

	s.mx.Lock()
	for _, user := range s.users {
		if user.refreshToken == req.RefreshToken {
			fmt.Printf("user refreshToken = %s\n, req token = %s", user.refreshToken, req.RefreshToken)
			resultUser = user
			break
		}
	}
	if resultUser == nil {
		err = status.New(codes.Unauthenticated, "refresh token is not valid").Err()
	}
	if err == nil {
		resultUser.accessToken = uuid.New().String()
		resultUser.refreshToken = uuid.New().String()
	}
	defer s.mx.Unlock()

	if err != nil {
		return nil, err
	}

	return &pb.RefreshResponse{
		AccessToken:  resultUser.accessToken,
		RefreshToken: resultUser.refreshToken,
		UserId:       resultUser.id,
	}, nil
}

func main() {
	implementation := NewServer() // наша реализация сервера

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, implementation) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	// Register:
	// grpc_cli call --json_input --json_output localhost:8082 AuthService/Register '{"email":"stas@gmail.com", "password":"123456"}'
	// Login:
	// grpc_cli call --json_input --json_output localhost:8082 AuthService/Login '{"email":"stas@gmail.com", "password":"123456"}'
	// Refresh:
	// grpc_cli call --json_input --json_output localhost:8082 AuthService/Refresh '{"refreshToken":"ff9bf764-0bfc-4e5d-b59f-cdaeeececb06"}'
}
