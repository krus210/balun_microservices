package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"gateway/pkg/api/auth"
	"gateway/pkg/api/chat"
	pb "gateway/pkg/api/gateway"
	"gateway/pkg/api/social"
	"gateway/pkg/api/users"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedGatewayServiceServer

	authClient   auth.AuthServiceClient
	usersClient  users.UsersServiceClient
	socialClient social.SocialServiceClient
	chatClient   chat.ChatServiceClient
}

func NewServer() (*Server, error) {
	authConn, err := grpc.NewClient("auth:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	usersConn, err := grpc.NewClient("users:8082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to users service: %w", err)
	}

	socialConn, err := grpc.NewClient("social:8083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to social service: %w", err)
	}

	chatConn, err := grpc.NewClient("chat:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chat service: %w", err)
	}

	srv := &Server{
		authClient:   auth.NewAuthServiceClient(authConn),
		usersClient:  users.NewUsersServiceClient(usersConn),
		socialClient: social.NewSocialServiceClient(socialConn),
		chatClient:   chat.NewChatServiceClient(chatConn),
	}

	return srv, nil
}

// customHTTPError обрабатывает gRPC ошибки и возвращает соответствующие HTTP коды
func customHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", marshaler.ContentType("application/json"))

	// Извлекаем gRPC статус из ошибки
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	// Мапим gRPC код в HTTP статус
	httpStatus := runtime.HTTPStatusFromCode(s.Code())
	w.WriteHeader(httpStatus)

	// Формируем JSON ответ
	type errorResponse struct {
		Error   string `json:"error"`
		Code    int32  `json:"code"`
		Message string `json:"message"`
	}

	resp := &errorResponse{
		Error:   http.StatusText(httpStatus),
		Code:    int32(s.Code()),
		Message: s.Message(),
	}

	// Маршалим ответ
	buf, merr := marshaler.Marshal(resp)
	if merr != nil {
		log.Printf("Failed to marshal error response: %v", merr)
		w.Write([]byte(`{"error":"Internal Server Error","code":13,"message":"failed to marshal error message"}`))
		return
	}

	if _, werr := w.Write(buf); werr != nil {
		log.Printf("Failed to write error response: %v", werr)
	}
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	log.Printf("Gateway: Register request for email: %s", req.GetEmail())

	resp, err := s.authClient.Register(ctx, req)
	if err != nil {
		log.Printf("Gateway: Register error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	log.Printf("Gateway: Login request for email: %s", req.GetEmail())

	resp, err := s.authClient.Login(ctx, req)
	if err != nil {
		log.Printf("Gateway: Login error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	log.Printf("Gateway: Refresh request")

	resp, err := s.authClient.Refresh(ctx, req)
	if err != nil {
		log.Printf("Gateway: Refresh error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) CreateProfile(ctx context.Context, req *users.CreateProfileRequest) (*users.CreateProfileResponse, error) {
	log.Printf("Gateway: CreateProfile request for user: %s", req.GetNickname())

	resp, err := s.usersClient.CreateProfile(ctx, req)
	if err != nil {
		log.Printf("Gateway: CreateProfile error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *users.UpdateProfileRequest) (*users.UpdateProfileResponse, error) {
	log.Printf("Gateway: UpdateProfile request for userId: %d", req.GetUserId())

	resp, err := s.usersClient.UpdateProfile(ctx, req)
	if err != nil {
		log.Printf("Gateway: UpdateProfile error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetProfileByID(ctx context.Context, req *users.GetProfileByIDRequest) (*users.GetProfileByIDResponse, error) {
	log.Printf("Gateway: GetProfileByID request for userId: %d", req.GetUserId())

	resp, err := s.usersClient.GetProfileByID(ctx, req)
	if err != nil {
		log.Printf("Gateway: GetProfileByID error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetProfileByNickname(ctx context.Context, req *users.GetProfileByNicknameRequest) (*users.GetProfileByNicknameResponse, error) {
	log.Printf("Gateway: GetProfileByNickname request for nickname: %s", req.GetNickname())

	resp, err := s.usersClient.GetProfileByNickname(ctx, req)
	if err != nil {
		log.Printf("Gateway: GetProfileByNickname error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) SearchByNickname(ctx context.Context, req *users.SearchByNicknameRequest) (*users.SearchByNicknameResponse, error) {
	log.Printf("Gateway: SearchByNickname request for query: %s", req.GetQuery())

	resp, err := s.usersClient.SearchByNickname(ctx, req)
	if err != nil {
		log.Printf("Gateway: SearchByNickname error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) SendFriendRequest(ctx context.Context, req *social.SendFriendRequestRequest) (*social.SendFriendRequestResponse, error) {
	log.Printf("Gateway: SendFriendRequest from userId: %d", req.GetUserId())

	resp, err := s.socialClient.SendFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: SendFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListRequests(ctx context.Context, req *social.ListRequestsRequest) (*social.ListRequestsResponse, error) {
	log.Printf("Gateway: ListRequests for userId: %d", req.GetUserId())

	resp, err := s.socialClient.ListRequests(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListRequests error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) AcceptFriendRequest(ctx context.Context, req *social.AcceptFriendRequestRequest) (*social.AcceptFriendRequestResponse, error) {
	log.Printf("Gateway: AcceptFriendRequest requestId: %d", req.GetRequestId())

	resp, err := s.socialClient.AcceptFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: AcceptFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeclineFriendRequest(ctx context.Context, req *social.DeclineFriendRequestRequest) (*social.DeclineFriendRequestResponse, error) {
	log.Printf("Gateway: DeclineFriendRequest requestId: %d", req.GetRequestId())

	resp, err := s.socialClient.DeclineFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: DeclineFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) RemoveFriend(ctx context.Context, req *social.RemoveFriendRequest) (*social.RemoveFriendResponse, error) {
	log.Printf("Gateway: RemoveFriend userId: %d", req.GetUserId())

	resp, err := s.socialClient.RemoveFriend(ctx, req)
	if err != nil {
		log.Printf("Gateway: RemoveFriend error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListFriends(ctx context.Context, req *social.ListFriendsRequest) (*social.ListFriendsResponse, error) {
	log.Printf("Gateway: ListFriends for userId: %d", req.GetUserId())

	resp, err := s.socialClient.ListFriends(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListFriends error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) CreateDirectChat(ctx context.Context, req *chat.CreateDirectChatRequest) (*chat.CreateDirectChatResponse, error) {
	log.Printf("Gateway: CreateDirectChat for participantId: %d", req.GetParticipantId())

	resp, err := s.chatClient.CreateDirectChat(ctx, req)
	if err != nil {
		log.Printf("Gateway: CreateDirectChat error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetChat(ctx context.Context, req *chat.GetChatRequest) (*chat.GetChatResponse, error) {
	log.Printf("Gateway: GetChat chatId: %d", req.GetChatId())

	resp, err := s.chatClient.GetChat(ctx, req)
	if err != nil {
		log.Printf("Gateway: GetChat error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListUserChats(ctx context.Context, req *chat.ListUserChatsRequest) (*chat.ListUserChatsResponse, error) {
	log.Printf("Gateway: ListUserChats for userId: %d", req.GetUserId())

	resp, err := s.chatClient.ListUserChats(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListUserChats error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListChatMembers(ctx context.Context, req *chat.ListChatMembersRequest) (*chat.ListChatMembersResponse, error) {
	log.Printf("Gateway: ListChatMembers for chatId: %d", req.GetChatId())

	resp, err := s.chatClient.ListChatMembers(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListChatMembers error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	log.Printf("Gateway: SendMessage in chatId: %d, text: %s", req.GetChatId(), req.GetText())

	resp, err := s.chatClient.SendMessage(ctx, req)
	if err != nil {
		log.Printf("Gateway: SendMessage error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListMessages(ctx context.Context, req *chat.ListMessagesRequest) (*chat.ListMessagesResponse, error) {
	log.Printf("Gateway: ListMessages for chatId: %d", req.GetChatId())

	resp, err := s.chatClient.ListMessages(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListMessages error: %v", err)
		return nil, err
	}

	return resp, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("failed to create Server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer()
		pb.RegisterGatewayServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8085")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("Gateway gRPC Server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Создаем ServeMux с кастомным обработчиком ошибок
		mux := runtime.NewServeMux(
			runtime.WithErrorHandler(customHTTPError),
		)
		if err = pb.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
			log.Fatalf("failed to register gateway handler: %v", err)
		}

		httpServer := &http.Server{Handler: mux}

		lis, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("Gateway HTTP Server listening at %v", lis.Addr())
		if err := httpServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Wait()
}
