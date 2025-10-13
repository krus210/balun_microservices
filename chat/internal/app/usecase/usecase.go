package usecase

import (
	"context"

	"chat/internal/app/models"
	"chat/internal/app/usecase/dto"
)

// Порты вторичные
type (
	UsersService interface {
		CheckUserExists(ctx context.Context, id models.UserID) (bool, error)
	}

	ChatRepository interface {
		SaveChat(ctx context.Context, chat *models.Chat) (*models.Chat, error)
		GetChat(ctx context.Context, chatID models.ChatID) (*models.Chat, error)
		GetDirectChatByParticipants(ctx context.Context, userID1, userID2 models.UserID) (*models.Chat, error)
		ListChatsByUserID(ctx context.Context, userID models.UserID) ([]*models.Chat, error)
		GetChatMembers(ctx context.Context, chatID models.ChatID) ([]models.UserID, error)
		IsChatMember(ctx context.Context, chatID models.ChatID, userID models.UserID) (bool, error)

		SaveMessage(ctx context.Context, msg *models.Message) (*models.Message, error)
		ListMessages(ctx context.Context, chatID models.ChatID, limit int64, cursor *string) (messages []*models.Message, nextCursor *string, err error)
	}
)

type Usecase interface {
	// CreateDirectChat создание личного чата
	CreateDirectChat(ctx context.Context, req dto.CreateDirectChatDto) (*models.Chat, error)
	// GetChat получение информации о чате
	GetChat(ctx context.Context, req dto.GetChatDto) (*models.Chat, error)
	// ListUserChats получение списка чатов пользователя
	ListUserChats(ctx context.Context, req dto.ListUserChatsDto) ([]*models.Chat, error)
	// ListChatMembers получение участников чата
	ListChatMembers(ctx context.Context, req dto.ListChatMembersDto) ([]models.UserID, error)
	// SendMessage отправка сообщения
	SendMessage(ctx context.Context, req dto.SendMessageDto) (*models.Message, error)
	// ListMessages получение истории сообщений
	ListMessages(ctx context.Context, req dto.ListMessagesDto) (*dto.ListMessagesResponse, error)
	// StreamMessages серверный стрим новых сообщений (будет реализован позже)
	// StreamMessages(ctx context.Context, req dto.StreamMessagesDto) (<-chan *models.Message, error)
}

type ChatService struct {
	usersService UsersService
	chatRepo     ChatRepository
}

var _ Usecase = (*ChatService)(nil)

func NewUsecase(usersService UsersService, chatRepo ChatRepository) *ChatService {
	return &ChatService{
		usersService: usersService,
		chatRepo:     chatRepo,
	}
}
