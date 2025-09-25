package main

import (
	"context"
	"log"
	"net"
	"sync"

	"buf.build/go/protovalidate"
	"github.com/AlekSi/pointer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "social/pkg/api"
)

type requestId int64

type Server struct {
	pb.SocialServiceServer

	mx        sync.RWMutex
	requests  map[requestId]*pb.FriendRequest
	nextReqId requestId
}

func NewServer() *Server {
	return &Server{
		requests:  make(map[requestId]*pb.FriendRequest),
		nextReqId: 1,
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

func (s *Server) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	// Проверяем, не отправляет ли пользователь заявку самому себе
	if req.GetUserId() == 0 {
		return nil, status.New(codes.InvalidArgument, "invalid user id").Err()
	}

	// Проверяем, не существует ли уже активная заявка
	for _, request := range s.requests {
		if request.UserId == req.GetUserId() && request.Status == pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING {
			return nil, status.New(codes.AlreadyExists, "friend request already exists").Err()
		}
	}

	// Создаем новую заявку
	friendRequest := &pb.FriendRequest{
		RequestId: int64(s.nextReqId),
		UserId:    req.GetUserId(),
		Status:    pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING,
	}

	s.requests[s.nextReqId] = friendRequest
	s.nextReqId++

	return &pb.SendFriendRequestResponse{FriendRequest: friendRequest}, nil
}

func (s *Server) ListRequests(ctx context.Context, _ *pb.ListRequestsRequest) (*pb.ListRequestsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	var pendingRequests []*pb.FriendRequest
	for _, request := range s.requests {
		if request.Status == pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING {
			pendingRequests = append(pendingRequests, request)
		}
	}

	return &pb.ListRequestsResponse{Requests: pendingRequests}, nil
}

func (s *Server) AcceptFriendRequest(ctx context.Context, req *pb.AcceptFriendRequestRequest) (*pb.AcceptFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	request, exists := s.requests[requestId(req.GetRequestId())]
	if !exists {
		return nil, status.New(codes.NotFound, "friend request not found").Err()
	}

	if request.Status != pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING {
		return nil, status.New(codes.PermissionDenied, "friend request is not pending").Err()
	}

	// Обновляем статус заявки
	request.Status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_ACCEPTED

	return &pb.AcceptFriendRequestResponse{FriendRequest: request}, nil
}

func (s *Server) DeclineFriendRequest(ctx context.Context, req *pb.DeclineFriendRequestRequest) (*pb.DeclineFriendRequestResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	request, exists := s.requests[requestId(req.GetRequestId())]
	if !exists {
		return nil, status.New(codes.NotFound, "friend request not found").Err()
	}

	if request.Status != pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_PENDING {
		return nil, status.New(codes.PermissionDenied, "friend request is not pending").Err()
	}

	// Обновляем статус заявки
	request.Status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_DECLINED

	return &pb.DeclineFriendRequestResponse{FriendRequest: request}, nil
}

func (s *Server) RemoveFriend(ctx context.Context, req *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.Lock()
	defer s.mx.Unlock()

	var resultRequest *pb.FriendRequest
	for _, request := range s.requests {
		if request.UserId == req.GetUserId() && request.Status == pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_ACCEPTED {
			resultRequest = request
		}
	}

	if resultRequest == nil {
		return nil, status.New(codes.NotFound, "user not found").Err()
	}

	resultRequest.Status = pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_DECLINED

	return &pb.RemoveFriendResponse{}, nil
}

func (s *Server) ListFriends(ctx context.Context, req *pb.ListFriendsRequest) (*pb.ListFriendsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	s.mx.RLock()
	defer s.mx.RUnlock()

	resultRequests := make([]int64, 0)
	for _, request := range s.requests {
		if request.UserId == req.GetUserId() && request.Status == pb.FriendRequestStatus_FRIEND_REQUEST_STATUS_ACCEPTED {
			resultRequests = append(resultRequests, request.RequestId)
			if req.GetLimit() > 0 && int64(len(resultRequests)) >= req.GetLimit() {
				break
			}
		}
	}

	return &pb.ListFriendsResponse{
		FriendUserIds: resultRequests,
		// TODO
		NextCursor: pointer.To("next"),
	}, nil
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
	pb.RegisterSocialServiceServer(server, implementation)

	reflection.Register(server)

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
