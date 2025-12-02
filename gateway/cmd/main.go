package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	"google.golang.org/grpc"

	"github.com/sskorolev/balun_microservices/lib/app"
	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"

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
		logger.ErrorKV(ctx, "failed to marshal error response", "error", merr.Error())
		w.Write([]byte(`{"error":"Internal Server Error","code":13,"message":"failed to marshal error message"}`))
		return
	}

	if _, werr := w.Write(buf); werr != nil {
		logger.ErrorKV(ctx, "failed to write error response", "error", werr.Error())
	}
}

func (s *Server) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	logger.InfoKV(ctx, "Gateway: Register request", "email", req.GetEmail())

	resp, err := s.authClient.Register(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: Register error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	logger.InfoKV(ctx, "Gateway: Login request", "email", req.GetEmail())

	resp, err := s.authClient.Login(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: Login error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	logger.InfoKV(ctx, "Gateway: Refresh request")

	resp, err := s.authClient.Refresh(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: Refresh error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) Logout(ctx context.Context, req *auth.LogoutRequest) (*auth.LogoutResponse, error) {
	logger.InfoKV(ctx, "Gateway: Logout request")

	resp, err := s.authClient.Logout(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: Logout error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetJWKS(ctx context.Context, req *auth.GetJWKSRequest) (*auth.GetJWKSResponse, error) {
	logger.InfoKV(ctx, "Gateway: GetJWKS request")

	resp, err := s.authClient.GetJWKS(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: GetJWKS error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) CreateProfile(ctx context.Context, req *users.CreateProfileRequest) (*users.CreateProfileResponse, error) {
	logger.InfoKV(ctx, "Gateway: CreateProfile request", "nickname", req.GetNickname())

	resp, err := s.usersClient.CreateProfile(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: CreateProfile error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *users.UpdateProfileRequest) (*users.UpdateProfileResponse, error) {
	logger.InfoKV(ctx, "Gateway: UpdateProfile request", "user_id", req.GetUserId())

	resp, err := s.usersClient.UpdateProfile(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: UpdateProfile error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetProfileByID(ctx context.Context, req *users.GetProfileByIDRequest) (*users.GetProfileByIDResponse, error) {
	logger.InfoKV(ctx, "Gateway: GetProfileByID request", "user_id", req.GetUserId())

	resp, err := s.usersClient.GetProfileByID(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: GetProfileByID error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetProfileByNickname(ctx context.Context, req *users.GetProfileByNicknameRequest) (*users.GetProfileByNicknameResponse, error) {
	logger.InfoKV(ctx, "Gateway: GetProfileByNickname request", "nickname", req.GetNickname())

	resp, err := s.usersClient.GetProfileByNickname(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: GetProfileByNickname error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) SearchByNickname(ctx context.Context, req *users.SearchByNicknameRequest) (*users.SearchByNicknameResponse, error) {
	logger.InfoKV(ctx, "Gateway: SearchByNickname request", "query", req.GetQuery())

	resp, err := s.usersClient.SearchByNickname(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: SearchByNickname error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) SendFriendRequest(ctx context.Context, req *social.SendFriendRequestRequest) (*social.SendFriendRequestResponse, error) {
	logger.InfoKV(ctx, "Gateway: SendFriendRequest", "to_user_id", req.GetToUserId())

	resp, err := s.socialClient.SendFriendRequest(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: SendFriendRequest error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListRequests(ctx context.Context, req *social.ListRequestsRequest) (*social.ListRequestsResponse, error) {
	logger.InfoKV(ctx, "Gateway: ListRequests", "to_user_id", req.GetToUserId())

	resp, err := s.socialClient.ListRequests(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: ListRequests error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) AcceptFriendRequest(ctx context.Context, req *social.AcceptFriendRequestRequest) (*social.AcceptFriendRequestResponse, error) {
	logger.InfoKV(ctx, "Gateway: AcceptFriendRequest", "request_id", req.GetRequestId())

	resp, err := s.socialClient.AcceptFriendRequest(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: AcceptFriendRequest error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeclineFriendRequest(ctx context.Context, req *social.DeclineFriendRequestRequest) (*social.DeclineFriendRequestResponse, error) {
	logger.InfoKV(ctx, "Gateway: DeclineFriendRequest", "request_id", req.GetRequestId())

	resp, err := s.socialClient.DeclineFriendRequest(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: DeclineFriendRequest error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) RemoveFriend(ctx context.Context, req *social.RemoveFriendRequest) (*social.RemoveFriendResponse, error) {
	logger.InfoKV(ctx, "Gateway: RemoveFriend", "user_id", req.GetUserId())

	resp, err := s.socialClient.RemoveFriend(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: RemoveFriend error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListFriends(ctx context.Context, req *social.ListFriendsRequest) (*social.ListFriendsResponse, error) {
	logger.InfoKV(ctx, "Gateway: ListFriends", "user_id", req.GetUserId())

	resp, err := s.socialClient.ListFriends(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: ListFriends error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) CreateDirectChat(ctx context.Context, req *chat.CreateDirectChatRequest) (*chat.CreateDirectChatResponse, error) {
	logger.InfoKV(ctx, "Gateway: CreateDirectChat", "participant_id", req.GetParticipantId())

	// Извлекаем Idempotency-Key из входящих метаданных
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if idempotencyKeys := md.Get("idempotency-key"); len(idempotencyKeys) > 0 {
			// Добавляем idempotency-key в исходящие метаданные
			ctx = metadata.AppendToOutgoingContext(ctx, "idempotency-key", idempotencyKeys[0])
			logger.InfoKV(ctx, "Gateway: Forwarding Idempotency-Key", "key", idempotencyKeys[0])
		}
	}

	resp, err := s.chatClient.CreateDirectChat(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: CreateDirectChat error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetChat(ctx context.Context, req *chat.GetChatRequest) (*chat.GetChatResponse, error) {
	logger.InfoKV(ctx, "Gateway: GetChat", "chat_id", req.GetChatId())

	resp, err := s.chatClient.GetChat(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: GetChat error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListUserChats(ctx context.Context, req *chat.ListUserChatsRequest) (*chat.ListUserChatsResponse, error) {
	logger.InfoKV(ctx, "Gateway: ListUserChats", "user_id", req.GetUserId())

	resp, err := s.chatClient.ListUserChats(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: ListUserChats error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListChatMembers(ctx context.Context, req *chat.ListChatMembersRequest) (*chat.ListChatMembersResponse, error) {
	logger.InfoKV(ctx, "Gateway: ListChatMembers", "chat_id", req.GetChatId())

	resp, err := s.chatClient.ListChatMembers(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: ListChatMembers error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.SendMessageResponse, error) {
	logger.InfoKV(ctx, "Gateway: SendMessage", "chat_id", req.GetChatId(), "text", req.GetText())

	resp, err := s.chatClient.SendMessage(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: SendMessage error", "error", err.Error())
		return nil, err
	}

	return resp, nil
}

func (s *Server) ListMessages(ctx context.Context, req *chat.ListMessagesRequest) (*chat.ListMessagesResponse, error) {
	logger.InfoKV(ctx, "Gateway: ListMessages", "chat_id", req.GetChatId())

	resp, err := s.chatClient.ListMessages(ctx, req)
	if err != nil {
		logger.ErrorKV(ctx, "Gateway: ListMessages error", "error", err.Error())
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
		logger.FatalKV(ctx, "failed to load config", "error", err.Error())
	}

	// Создаем приложение
	application, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.FatalKV(ctx, "failed to create app", "error", err.Error())
	}

	// Инициализируем logger
	if err := application.InitLogger(cfg.Logger, cfg.Service.Name, cfg.Service.Environment); err != nil {
		logger.FatalKV(ctx, "failed to initialize logger", "error", err.Error())
	}

	// Инициализируем tracer
	if err := application.InitTracer(cfg.Tracer); err != nil {
		logger.FatalKV(ctx, "failed to initialize tracer", "error", err.Error())
	}

	// Инициализируем metrics
	if err := application.InitMetrics(cfg.Metrics, cfg.Service.Name); err != nil {
		logger.FatalKV(ctx, "failed to initialize metrics", "error", err.Error())
	}

	// Инициализируем admin HTTP сервер (метрики и pprof)
	if cfg.Server.Admin != nil {
		if err := application.InitAdminServer(*cfg.Server.Admin); err != nil {
			logger.FatalKV(ctx, "failed to initialize admin server", "error", err.Error())
		}
	}

	// Инициализируем gRPC клиенты для всех сервисов
	if err := application.InitGRPCClient(ctx, "auth", cfg.AuthService); err != nil {
		logger.FatalKV(ctx, "failed to connect to auth service", "error", err.Error())
	}

	if err := application.InitGRPCClient(ctx, "users", cfg.UsersService); err != nil {
		logger.FatalKV(ctx, "failed to connect to users service", "error", err.Error())
	}

	if err := application.InitGRPCClient(ctx, "social", cfg.SocialService); err != nil {
		logger.FatalKV(ctx, "failed to connect to social service", "error", err.Error())
	}

	if err := application.InitGRPCClient(ctx, "chat", cfg.ChatService); err != nil {
		logger.FatalKV(ctx, "failed to connect to chat service", "error", err.Error())
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
		logger.FatalKV(ctx, "failed to register gateway handler", "error", err.Error())
	}

	// Инициализируем HTTP handler
	application.InitHTTPServer(mux)

	// Запускаем все три сервера через новый метод
	logger.InfoKV(ctx, "starting gateway service",
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
		"http_port", cfg.Server.HTTP.Port,
		"grpc_port", cfg.Server.GRPC.Port,
		"admin_port", cfg.Server.Admin.Port,
	)

	err = application.RunServers(ctx, app.ServerConfig{
		GRPC:  cfg.Server.GRPC,
		HTTP:  cfg.Server.HTTP,
		Admin: cfg.Server.Admin,
	})

	switch {
	case err == nil || errors.Is(err, context.Canceled):
		logger.InfoKV(ctx, "gateway service components stopped")
	case errors.Is(err, context.DeadlineExceeded):
		logger.WarnKV(ctx, "graceful shutdown timeout exceeded, forcing cleanup")
	default:
		logger.FatalKV(ctx, "failed to serve", "error", err.Error())
	}

	logger.InfoKV(ctx, "gateway service shutdown complete")
}
