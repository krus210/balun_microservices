package grpc

import (
	"context"
	"log"
	"sync"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"

	pb "chat/pkg/api"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	idempotencyKeys  = make(map[string]bool)
	idempotencyMutex sync.RWMutex
)

func (h *ChatController) CreateDirectChat(ctx context.Context, req *pb.CreateDirectChatRequest) (*pb.CreateDirectChatResponse, error) {
	// Валидируем и сохраняем idempotency-key
	if err := validateIdempotencyKey(ctx); err != nil {
		return nil, err
	}

	chat, err := h.usecase.CreateDirectChat(ctx, dto.CreateDirectChatDto{
		UserID:        "1", // TODO: брать из хедера
		ParticipantID: models.UserID(req.ParticipantId),
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateDirectChatResponse{
		ChatId: string(chat.ID),
	}, nil
}

// validateIdempotencyKey извлекает, валидирует idempotency-key и сохраняет его в мапу
func validateIdempotencyKey(ctx context.Context) error {
	// Извлекаем metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// Метаданных нет вообще
		return buildBadRequestError("idempotency-key обязателен", "idempotency-key", "missing")
	}

	// Получаем ключ
	keys := md.Get("idempotency-key")
	if len(keys) == 0 || keys[0] == "" {
		// Ключ отсутствует или пустой
		return buildBadRequestError("idempotency-key обязателен", "idempotency-key", "empty")
	}

	key := keys[0]
	log.Printf("Received Idempotency-Key: %s", key)

	// Проверяем, использовался ли уже этот ключ
	idempotencyMutex.Lock()
	defer idempotencyMutex.Unlock()

	if idempotencyKeys[key] {
		log.Printf("Duplicate request with Idempotency-Key: %s", key)
		return buildBadRequestError("запрос с таким idempotency-key уже был обработан", "idempotency-key", "duplicate request")
	}

	// Сохраняем ключ как использованный
	idempotencyKeys[key] = true
	return nil
}

// buildBadRequestError создает ошибку BadRequest с детальным описанием
func buildBadRequestError(message, field, description string) error {
	rpcErr := status.New(codes.InvalidArgument, message)
	detailedError, err := rpcErr.WithDetails(&errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{
				Field:       field,
				Description: description,
			},
		},
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return detailedError.Err()
}
