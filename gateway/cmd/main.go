package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"

	"gateway/pkg/api/auth"
	"gateway/pkg/api/chat"
	pb "gateway/pkg/api/gateway"
	"gateway/pkg/api/social"
	"gateway/pkg/api/users"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedGatewayServiceServer

	authClient   auth.AuthServiceClient
	usersClient  users.UsersServiceClient
	socialClient social.SocialServiceClient
	chatClient   chat.ChatServiceClient
}

func NewServer(application *app.App) *Server {
	return &Server{
		authClient:   auth.NewAuthServiceClient(application.GetGRPCClient("auth")),
		usersClient:  users.NewUsersServiceClient(application.GetGRPCClient("users")),
		socialClient: social.NewSocialServiceClient(application.GetGRPCClient("social")),
		chatClient:   chat.NewChatServiceClient(application.GetGRPCClient("chat")),
	}
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
	log.Printf("Gateway: UpdateProfile request for userId: %s", req.GetUserId())

	resp, err := s.usersClient.UpdateProfile(ctx, req)
	if err != nil {
		log.Printf("Gateway: UpdateProfile error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetProfileByID(ctx context.Context, req *users.GetProfileByIDRequest) (*users.GetProfileByIDResponse, error) {
	log.Printf("Gateway: GetProfileByID request for userId: %s", req.GetUserId())

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
	log.Printf("Gateway: SendFriendRequest from userId: %s", req.GetToUserId())

	resp, err := s.socialClient.SendFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: SendFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListRequests(ctx context.Context, req *social.ListRequestsRequest) (*social.ListRequestsResponse, error) {
	log.Printf("Gateway: ListRequests for userId: %s", req.GetToUserId())

	resp, err := s.socialClient.ListRequests(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListRequests error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) AcceptFriendRequest(ctx context.Context, req *social.AcceptFriendRequestRequest) (*social.AcceptFriendRequestResponse, error) {
	log.Printf("Gateway: AcceptFriendRequest requestId: %s", req.GetRequestId())

	resp, err := s.socialClient.AcceptFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: AcceptFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeclineFriendRequest(ctx context.Context, req *social.DeclineFriendRequestRequest) (*social.DeclineFriendRequestResponse, error) {
	log.Printf("Gateway: DeclineFriendRequest requestId: %s", req.GetRequestId())

	resp, err := s.socialClient.DeclineFriendRequest(ctx, req)
	if err != nil {
		log.Printf("Gateway: DeclineFriendRequest error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) RemoveFriend(ctx context.Context, req *social.RemoveFriendRequest) (*social.RemoveFriendResponse, error) {
	log.Printf("Gateway: RemoveFriend userId: %s", req.GetUserId())

	resp, err := s.socialClient.RemoveFriend(ctx, req)
	if err != nil {
		log.Printf("Gateway: RemoveFriend error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListFriends(ctx context.Context, req *social.ListFriendsRequest) (*social.ListFriendsResponse, error) {
	log.Printf("Gateway: ListFriends for userId: %s", req.GetUserId())

	resp, err := s.socialClient.ListFriends(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListFriends error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) CreateDirectChat(ctx context.Context, req *chat.CreateDirectChatRequest) (*chat.CreateDirectChatResponse, error) {
	log.Printf("Gateway: CreateDirectChat for participantId: %s", req.GetParticipantId())

	// Извлекаем Idempotency-Key из входящих метаданных
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if idempotencyKeys := md.Get("idempotency-key"); len(idempotencyKeys) > 0 {
			// Добавляем idempotency-key в исходящие метаданные
			ctx = metadata.AppendToOutgoingContext(ctx, "idempotency-key", idempotencyKeys[0])
			log.Printf("Gateway: Forwarding Idempotency-Key: %s", idempotencyKeys[0])
		}
	}

	resp, err := s.chatClient.CreateDirectChat(ctx, req)
	if err != nil {
		log.Printf("Gateway: CreateDirectChat error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetChat(ctx context.Context, req *chat.GetChatRequest) (*chat.GetChatResponse, error) {
	log.Printf("Gateway: GetChat chatId: %s", req.GetChatId())

	resp, err := s.chatClient.GetChat(ctx, req)
	if err != nil {
		log.Printf("Gateway: GetChat error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListUserChats(ctx context.Context, req *chat.ListUserChatsRequest) (*chat.ListUserChatsResponse, error) {
	log.Printf("Gateway: ListUserChats for userId: %s", req.GetUserId())

	resp, err := s.chatClient.ListUserChats(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListUserChats error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListChatMembers(ctx context.Context, req *chat.ListChatMembersRequest) (*chat.ListChatMembersResponse, error) {
	log.Printf("Gateway: ListChatMembers for chatId: %s", req.GetChatId())

	resp, err := s.chatClient.ListChatMembers(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListChatMembers error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	log.Printf("Gateway: SendMessage in chatId: %s, text: %s", req.GetChatId(), req.GetText())

	resp, err := s.chatClient.SendMessage(ctx, req)
	if err != nil {
		log.Printf("Gateway: SendMessage error: %v", err)
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListMessages(ctx context.Context, req *chat.ListMessagesRequest) (*chat.ListMessagesResponse, error) {
	log.Printf("Gateway: ListMessages for chatId: %s", req.GetChatId())

	resp, err := s.chatClient.ListMessages(ctx, req)
	if err != nil {
		log.Printf("Gateway: ListMessages error: %v", err)
		return nil, err
	}

	return resp, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем конфигурацию через lib/config
	cfg, err := config.LoadServiceConfig(ctx, "gateway",
		config.WithoutDatabase(),
		config.WithAuthService("auth", 8082),
		config.WithUsersService("users", 8082),
		config.WithSocialService("social", 8082),
		config.WithChatService("chat", 8082),
	)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	// Инициализируем gRPC клиенты для всех сервисов
	if err := application.InitGRPCClient(ctx, "auth", cfg.AuthService); err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}

	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		log.Fatalf("failed to connect to users service: %v", err)
	}

	if err := application.InitGRPCClient(ctx, "social", cfg.SocialService); err != nil {
		log.Fatalf("failed to connect to social service: %v", err)
	}

	if err := application.InitGRPCClient(ctx, "chat", cfg.ChatService); err != nil {
		log.Fatalf("failed to connect to chat service: %v", err)
	}

	// Создаем Server с клиентами
	server := NewServer(application)

	// Инициализируем gRPC сервер
	application.InitGRPCServer(cfg.Server)

	// Регистрируем gRPC сервисы
	application.RegisterGRPC(func(s *grpc.Server) {
		pb.RegisterGatewayServiceServer(s, server)
	})

	// Создаем HTTP ServeMux с кастомным обработчиком ошибок
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(customHTTPError),
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			// Пробрасываем idempotency-key в metadata
			// HTTP заголовки приходят в разных регистрах
			switch key {
			case "Idempotency-Key", "idempotency-key", "X-Idempotency-Key", "x-idempotency-key":
				return "idempotency-key", true
			}
			// Стандартные заголовки (Authorization и т.д.)
			return runtime.DefaultHeaderMatcher(strings.ToLower(key))
		}),
	)
	if err := pb.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
		log.Fatalf("failed to register gateway handler: %v", err)
	}

	// Инициализируем HTTP handler
	application.InitHTTPServer(mux)

	// Запускаем оба сервера (HTTP и gRPC) параллельно
	slog.Info("starting gateway service", "http_port", cfg.Server.HTTP.Port, "grpc_port", cfg.Server.GRPC.Port)

	if err := application.RunBoth(ctx, *cfg.Server.GRPC, *cfg.Server.HTTP); err != nil {
		if err != context.Canceled {
			log.Fatalf("failed to serve: %v", err)
		}
	}
	slog.Info("gateway service stopped gracefully")
}
