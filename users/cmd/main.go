package main

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "users/pkg/api"
)

type serialId int64

type Server struct {
	pb.UsersServiceServer

	// Бизнес логика/зависимости
	mx    sync.RWMutex
	users map[serialId]*pb.UserProfile
}

func NewServer() *Server {
	return &Server{users: make(map[serialId]*pb.UserProfile)}
}

func validationInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to initialize validator")
	}

	if protoReq, ok := req.(proto.Message); ok {
		if err := validator.Validate(protoReq); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}

func (s *Server) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	var resultUser *pb.UserProfile
	s.mx.Lock()
	defer s.mx.Unlock()
	if user, ok := s.users[serialId(req.GetUserId())]; ok && user != nil {
		return nil, status.New(codes.AlreadyExists, "user already exists").Err()
	}

	resultUser = &pb.UserProfile{
		UserId:    req.GetUserId(),
		Nickname:  req.GetNickname(),
		Bio:       req.Bio,
		AvatarUrl: req.AvatarUrl,
	}
	s.users[serialId(req.GetUserId())] = resultUser

	return &pb.CreateProfileResponse{UserProfile: resultUser}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	if user, ok := s.users[serialId(req.GetUserId())]; !ok || user == nil {
		return nil, status.New(codes.NotFound, "user not found").Err()
	}

	user := s.users[serialId(req.GetUserId())]
	if req.GetNickname() != "" {
		user.Nickname = req.GetNickname()
	}
	if req.GetBio() != "" {
		user.Bio = req.Bio
	}
	if req.GetAvatarUrl() != "" {
		user.AvatarUrl = req.AvatarUrl
	}

	return &pb.UpdateProfileResponse{UserProfile: user}, nil
}

func (s *Server) GetProfileByID(ctx context.Context, req *pb.GetProfileByIDRequest) (*pb.GetProfileByIDResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	user, ok := s.users[serialId(req.GetUserId())]

	if !ok || user == nil {
		return nil, status.New(codes.NotFound, "user not found").Err()
	}

	return &pb.GetProfileByIDResponse{UserProfile: user}, nil
}

func (s *Server) GetProfileByNickname(ctx context.Context, req *pb.GetProfileByNicknameRequest) (*pb.GetProfileByNicknameResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	var resultUser *pb.UserProfile
	for _, user := range s.users {
		if user != nil && strings.EqualFold(user.Nickname, req.Nickname) {
			resultUser = user
			break
		}
	}

	if resultUser == nil {
		return nil, status.New(codes.NotFound, "user not found").Err()
	}

	return &pb.GetProfileByNicknameResponse{UserProfile: resultUser}, nil
}

func (s *Server) SearchByNickname(ctx context.Context, req *pb.SearchByNicknameRequest) (*pb.SearchByNicknameResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	resultUsers := make([]*pb.UserProfile, 0)
	for _, user := range s.users {
		if user != nil && strings.Contains(strings.ToLower(user.GetNickname()), strings.ToLower(req.GetQuery())) {
			resultUsers = append(resultUsers, user)
			if req.Limit != 0 && int64(len(resultUsers)) >= req.Limit {
				break
			}
		}
	}

	return &pb.SearchByNicknameResponse{Results: resultUsers}, nil
}

func main() {
	implementation := NewServer() // наша реализация сервера

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(validationInterceptor),
	)
	pb.RegisterUsersServiceServer(server, implementation) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
