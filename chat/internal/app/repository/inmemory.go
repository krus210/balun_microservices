package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"chat/internal/app/models"

	"github.com/google/uuid"
)

type InMemoryChatRepository struct {
	mu       sync.RWMutex
	chats    map[models.ChatID]*models.Chat
	messages map[models.ChatID][]*models.Message
}

func NewInMemoryChatRepository() *InMemoryChatRepository {
	return &InMemoryChatRepository{
		chats:    make(map[models.ChatID]*models.Chat),
		messages: make(map[models.ChatID][]*models.Message),
	}
}

func (r *InMemoryChatRepository) SaveChat(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if chat.ID == "" {
		chat.ID = models.ChatID(uuid.New().String())
		chat.CreatedAt = time.Now()
	}
	chat.UpdatedAt = time.Now()

	// Deep copy to avoid external modifications
	chatCopy := &models.Chat{
		ID:             chat.ID,
		ParticipantIDs: make([]models.UserID, len(chat.ParticipantIDs)),
		Messages:       make([]models.Message, len(chat.Messages)),
		CreatedAt:      chat.CreatedAt,
		UpdatedAt:      chat.UpdatedAt,
	}
	copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
	copy(chatCopy.Messages, chat.Messages)

	r.chats[chat.ID] = chatCopy
	return chat, nil
}

func (r *InMemoryChatRepository) GetChat(ctx context.Context, chatID models.ChatID) (*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return nil, nil
	}

	// Deep copy to avoid external modifications
	chatCopy := &models.Chat{
		ID:             chat.ID,
		ParticipantIDs: make([]models.UserID, len(chat.ParticipantIDs)),
		Messages:       make([]models.Message, len(chat.Messages)),
		CreatedAt:      chat.CreatedAt,
		UpdatedAt:      chat.UpdatedAt,
	}
	copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
	copy(chatCopy.Messages, chat.Messages)

	return chatCopy, nil
}

func (r *InMemoryChatRepository) GetDirectChatByParticipants(ctx context.Context, userID1, userID2 models.UserID) (*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, chat := range r.chats {
		if len(chat.ParticipantIDs) == 2 {
			hasUser1 := false
			hasUser2 := false
			for _, participantID := range chat.ParticipantIDs {
				if participantID == userID1 {
					hasUser1 = true
				}
				if participantID == userID2 {
					hasUser2 = true
				}
			}
			if hasUser1 && hasUser2 {
				// Deep copy
				chatCopy := &models.Chat{
					ID:             chat.ID,
					ParticipantIDs: make([]models.UserID, len(chat.ParticipantIDs)),
					Messages:       make([]models.Message, len(chat.Messages)),
					CreatedAt:      chat.CreatedAt,
					UpdatedAt:      chat.UpdatedAt,
				}
				copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
				copy(chatCopy.Messages, chat.Messages)
				return chatCopy, nil
			}
		}
	}

	return nil, nil
}

func (r *InMemoryChatRepository) ListChatsByUserID(ctx context.Context, userID models.UserID) ([]*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Chat
	for _, chat := range r.chats {
		for _, participantID := range chat.ParticipantIDs {
			if participantID == userID {
				// Deep copy
				chatCopy := &models.Chat{
					ID:             chat.ID,
					ParticipantIDs: make([]models.UserID, len(chat.ParticipantIDs)),
					Messages:       make([]models.Message, len(chat.Messages)),
					CreatedAt:      chat.CreatedAt,
					UpdatedAt:      chat.UpdatedAt,
				}
				copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
				copy(chatCopy.Messages, chat.Messages)
				result = append(result, chatCopy)
				break
			}
		}
	}

	return result, nil
}

func (r *InMemoryChatRepository) GetChatMembers(ctx context.Context, chatID models.ChatID) ([]models.UserID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return nil, nil
	}

	members := make([]models.UserID, len(chat.ParticipantIDs))
	copy(members, chat.ParticipantIDs)
	return members, nil
}

func (r *InMemoryChatRepository) IsChatMember(ctx context.Context, chatID models.ChatID, userID models.UserID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return false, nil
	}

	for _, participantID := range chat.ParticipantIDs {
		if participantID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (r *InMemoryChatRepository) SaveMessage(ctx context.Context, msg *models.Message) (*models.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check chat exists
	if _, exists := r.chats[msg.ChatID]; !exists {
		return nil, nil
	}

	if msg.ID == "" {
		msg.ID = models.MessageID(uuid.New().String())
		msg.CreatedAt = time.Now()
	}
	msg.UpdatedAt = time.Now()

	// Deep copy
	msgCopy := &models.Message{
		ID:        msg.ID,
		Text:      msg.Text,
		ChatID:    msg.ChatID,
		OwnerID:   msg.OwnerID,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: msg.UpdatedAt,
	}

	r.messages[msg.ChatID] = append(r.messages[msg.ChatID], msgCopy)
	return msg, nil
}

func (r *InMemoryChatRepository) ListMessages(ctx context.Context, chatID models.ChatID, limit int64, cursor *string) (messages []*models.Message, nextCursor *string, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check chat exists
	if _, exists := r.chats[chatID]; !exists {
		return nil, nil, nil
	}

	chatMessages := r.messages[chatID]
	if len(chatMessages) == 0 {
		return []*models.Message{}, nil, nil
	}

	// Sort messages by ID (ascending - oldest first)
	sortedMessages := make([]*models.Message, len(chatMessages))
	copy(sortedMessages, chatMessages)
	sort.Slice(sortedMessages, func(i, j int) bool {
		return sortedMessages[i].ID < sortedMessages[j].ID
	})

	// Find start position based on cursor
	startIdx := 0
	if cursor != nil && *cursor != "" {
		// Find the position after the cursor
		for i, msg := range sortedMessages {
			if msg.ID == models.MessageID(*cursor) {
				startIdx = i + 1
				break
			}
		}
	}

	// Collect messages up to limit
	var result []*models.Message
	endIdx := startIdx + int(limit)
	if endIdx > len(sortedMessages) {
		endIdx = len(sortedMessages)
	}

	for i := startIdx; i < endIdx; i++ {
		// Deep copy
		msgCopy := &models.Message{
			ID:        sortedMessages[i].ID,
			Text:      sortedMessages[i].Text,
			ChatID:    sortedMessages[i].ChatID,
			OwnerID:   sortedMessages[i].OwnerID,
			CreatedAt: sortedMessages[i].CreatedAt,
			UpdatedAt: sortedMessages[i].UpdatedAt,
		}
		result = append(result, msgCopy)
	}

	// Set next cursor if there are more messages
	if endIdx < len(sortedMessages) {
		lastID := string(result[len(result)-1].ID)
		nextCursor = &lastID
	}

	return result, nextCursor, nil
}
