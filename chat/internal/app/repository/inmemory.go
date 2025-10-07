package repository

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"chat/internal/app/models"
)

type InMemoryChatRepository struct {
	mu           sync.RWMutex
	chats        map[int64]*models.Chat
	messages     map[int64][]*models.Message
	chatIDSeq    int64
	messageIDSeq int64
}

func NewInMemoryChatRepository() *InMemoryChatRepository {
	return &InMemoryChatRepository{
		chats:    make(map[int64]*models.Chat),
		messages: make(map[int64][]*models.Message),
	}
}

func (r *InMemoryChatRepository) SaveChat(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if chat.ID == 0 {
		r.chatIDSeq++
		chat.ID = r.chatIDSeq
		chat.CreatedAt = time.Now()
	}
	chat.UpdatedAt = time.Now()

	// Deep copy to avoid external modifications
	chatCopy := &models.Chat{
		ID:             chat.ID,
		ParticipantIDs: make([]int64, len(chat.ParticipantIDs)),
		Messages:       make([]models.Message, len(chat.Messages)),
		CreatedAt:      chat.CreatedAt,
		UpdatedAt:      chat.UpdatedAt,
	}
	copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
	copy(chatCopy.Messages, chat.Messages)

	r.chats[chat.ID] = chatCopy
	return chat, nil
}

func (r *InMemoryChatRepository) GetChat(ctx context.Context, chatID int64) (*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return nil, fmt.Errorf("chat with id %d not found", chatID)
	}

	// Deep copy to avoid external modifications
	chatCopy := &models.Chat{
		ID:             chat.ID,
		ParticipantIDs: make([]int64, len(chat.ParticipantIDs)),
		Messages:       make([]models.Message, len(chat.Messages)),
		CreatedAt:      chat.CreatedAt,
		UpdatedAt:      chat.UpdatedAt,
	}
	copy(chatCopy.ParticipantIDs, chat.ParticipantIDs)
	copy(chatCopy.Messages, chat.Messages)

	return chatCopy, nil
}

func (r *InMemoryChatRepository) GetDirectChatByParticipants(ctx context.Context, userID1, userID2 int64) (*models.Chat, error) {
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
					ParticipantIDs: make([]int64, len(chat.ParticipantIDs)),
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

	return nil, fmt.Errorf("direct chat between users %d and %d not found", userID1, userID2)
}

func (r *InMemoryChatRepository) ListChatsByUserID(ctx context.Context, userID int64) ([]*models.Chat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Chat
	for _, chat := range r.chats {
		for _, participantID := range chat.ParticipantIDs {
			if participantID == userID {
				// Deep copy
				chatCopy := &models.Chat{
					ID:             chat.ID,
					ParticipantIDs: make([]int64, len(chat.ParticipantIDs)),
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

func (r *InMemoryChatRepository) GetChatMembers(ctx context.Context, chatID int64) ([]int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return nil, fmt.Errorf("chat with id %d not found", chatID)
	}

	members := make([]int64, len(chat.ParticipantIDs))
	copy(members, chat.ParticipantIDs)
	return members, nil
}

func (r *InMemoryChatRepository) IsChatMember(ctx context.Context, chatID int64, userID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chat, exists := r.chats[chatID]
	if !exists {
		return false, fmt.Errorf("chat with id %d not found", chatID)
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
		return nil, fmt.Errorf("chat with id %d not found", msg.ChatID)
	}

	if msg.ID == 0 {
		r.messageIDSeq++
		msg.ID = r.messageIDSeq
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

func (r *InMemoryChatRepository) ListMessages(ctx context.Context, chatID int64, limit int64, cursor *string) (messages []*models.Message, nextCursor *string, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check chat exists
	if _, exists := r.chats[chatID]; !exists {
		return nil, nil, fmt.Errorf("chat with id %d not found", chatID)
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
		cursorID, err := strconv.ParseInt(*cursor, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}

		// Find the position after the cursor
		for i, msg := range sortedMessages {
			if msg.ID == cursorID {
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
		lastID := strconv.FormatInt(result[len(result)-1].ID, 10)
		nextCursor = &lastID
	}

	return result, nextCursor, nil
}
