package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "chat/pkg/api"
)

type Server struct {
	pb.UnimplementedChatServiceServer

	mx         sync.RWMutex
	chats      map[int64]*pb.Chat
	messages   map[int64][]*pb.Message
	userChats  map[int64][]int64
	nextChatId int64
	nextMsgId  int64
}

func NewServer() *Server {
	return &Server{
		chats:      make(map[int64]*pb.Chat),
		messages:   make(map[int64][]*pb.Message),
		userChats:  make(map[int64][]int64),
		nextChatId: 1,
		nextMsgId:  1,
	}
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

func (s *Server) CreateDirectChat(ctx context.Context, req *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	if req.GetParticipantId() == 0 {
		return nil, status.New(codes.InvalidArgument, "invalid participant id").Err()
	}

	currentUserId := int64(1) // В реальности из контекста/токена
	if currentUserId == req.GetParticipantId() {
		return nil, status.New(codes.InvalidArgument, "cannot create chat with yourself").Err()
	}

	// Проверяем, не существует ли уже чат между этими пользователями
	for _, chat := range s.chats {
		if len(chat.ParticipantIds) == 2 {
			participants := make(map[int64]bool)
			for _, pid := range chat.ParticipantIds {
				participants[pid] = true
			}
			if participants[currentUserId] && participants[req.GetParticipantId()] {
				return nil, status.New(codes.AlreadyExists, "direct chat already exists").Err()
			}
		}
	}

	newChat := &pb.Chat{
		ChatId:            s.nextChatId,
		ParticipantIds:    []int64{currentUserId, req.GetParticipantId()},
		CreatedAtUnixMs:   time.Now().UnixMilli(),
		LastMessageUnixMs: nil,
	}

	s.chats[s.nextChatId] = newChat
	s.userChats[currentUserId] = append(s.userChats[currentUserId], s.nextChatId)
	s.userChats[req.GetParticipantId()] = append(s.userChats[req.GetParticipantId()], s.nextChatId)
	chatId := s.nextChatId
	s.nextChatId++

	return &pb.CreateDirectChatResponse{ChatId: chatId}, nil
}

func (s *Server) GetChat(ctx context.Context, req *pb.GetChatRequest) (*pb.GetChatResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	chat, exists := s.chats[req.GetChatId()]
	if !exists {
		return nil, status.New(codes.NotFound, "chat not found").Err()
	}

	// В реальности проверить, что пользователь имеет доступ к чату
	currentUserId := int64(1)
	hasAccess := false
	for _, pid := range chat.ParticipantIds {
		if pid == currentUserId {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		return nil, status.New(codes.PermissionDenied, "no access to chat").Err()
	}

	return &pb.GetChatResponse{Chat: chat}, nil
}

func (s *Server) ListUserChats(ctx context.Context, req *pb.ListUserChatsRequest) (*pb.ListUserChatsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	userChatIds := s.userChats[req.GetUserId()]
	var chats []*pb.Chat
	for _, cid := range userChatIds {
		if chat, exists := s.chats[cid]; exists {
			chats = append(chats, chat)
		}
	}

	return &pb.ListUserChatsResponse{Chats: chats}, nil
}

func (s *Server) ListChatMembers(ctx context.Context, req *pb.ListChatMembersRequest) (*pb.ListChatMembersResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	chat, exists := s.chats[req.GetChatId()]
	if !exists {
		return nil, status.New(codes.NotFound, "chat not found").Err()
	}

	return &pb.ListChatMembersResponse{UserIds: chat.ParticipantIds}, nil
}

func (s *Server) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	chat, exists := s.chats[req.GetChatId()]
	if !exists {
		return nil, status.New(codes.NotFound, "chat not found").Err()
	}

	if req.GetText() == "" {
		return nil, status.New(codes.InvalidArgument, "message text cannot be empty").Err()
	}

	currentUserId := int64(1)
	hasAccess := false
	for _, pid := range chat.ParticipantIds {
		if pid == currentUserId {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		return nil, status.New(codes.PermissionDenied, "no access to chat").Err()
	}

	now := time.Now().UnixMilli()
	message := &pb.Message{
		MessageId:    s.nextMsgId,
		ChatId:       req.GetChatId(),
		UserId:       currentUserId,
		Text:         req.GetText(),
		SentAtUnixMs: now,
	}

	s.messages[req.GetChatId()] = append(s.messages[req.GetChatId()], message)
	chat.LastMessageUnixMs = &now
	s.nextMsgId++

	return &pb.SendMessageResponse{Message: message}, nil
}

func (s *Server) ListMessages(ctx context.Context, req *pb.ListMessagesRequest) (*pb.ListMessagesResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	_, exists := s.chats[req.GetChatId()]
	if !exists {
		return nil, status.New(codes.NotFound, "chat not found").Err()
	}

	messages := s.messages[req.GetChatId()]
	limit := req.GetLimit()
	if limit <= 0 || limit > int64(len(messages)) {
		limit = int64(len(messages))
	}

	// Берем последние сообщения
	startIdx := int64(len(messages)) - limit
	if startIdx < 0 {
		startIdx = 0
	}

	resultMessages := messages[startIdx:]

	var nextCursor *string
	if len(messages) > int(limit) {
		cursor := "next_page"
		nextCursor = &cursor
	}

	return &pb.ListMessagesResponse{
		Messages:   resultMessages,
		NextCursor: nextCursor,
	}, nil
}

func (s *Server) StreamMessages(req *pb.StreamMessagesRequest, stream pb.ChatService_StreamMessagesServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	_, exists := s.chats[req.GetChatId()]
	s.mx.RUnlock()

	if !exists {
		return status.New(codes.NotFound, "chat not found").Err()
	}

	// Простая имитация стрима - отправляем существующие сообщения
	s.mx.RLock()
	messages := s.messages[req.GetChatId()]
	s.mx.RUnlock()

	sinceMs := req.GetSinceUnixMs()
	for _, msg := range messages {
		if sinceMs == 0 || msg.SentAtUnixMs > sinceMs {
			if err := stream.Send(&pb.StreamMessagesResponse{Message: msg}); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	implementation := NewServer()

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(validationInterceptor),
	)
	pb.RegisterChatServiceServer(server, implementation)

	reflection.Register(server)

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
