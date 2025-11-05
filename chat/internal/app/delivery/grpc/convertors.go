package grpc

import (
	"chat/internal/app/models"
	pb "chat/pkg/api"
)

func newPbChatFromChat(chat *models.Chat) *pb.Chat {
	var lastMessageUnixMs *int64
	if chat.UpdatedAt.Unix() > 0 {
		ms := chat.UpdatedAt.UnixMilli()
		lastMessageUnixMs = &ms
	}

	// Конвертация []models.UserID в []string
	participantIDs := make([]string, len(chat.ParticipantIDs))
	for i, id := range chat.ParticipantIDs {
		participantIDs[i] = string(id)
	}

	return &pb.Chat{
		ChatId:            string(chat.ID),
		ParticipantIds:    participantIDs,
		CreatedAtUnixMs:   chat.CreatedAt.UnixMilli(),
		LastMessageUnixMs: lastMessageUnixMs,
	}
}

func newPbChatsFromChats(chats []*models.Chat) []*pb.Chat {
	results := make([]*pb.Chat, len(chats))

	for i, chat := range chats {
		results[i] = newPbChatFromChat(chat)
	}

	return results
}

func newPbMessageFromMessage(msg *models.Message) *pb.Message {
	return &pb.Message{
		MessageId:    string(msg.ID),
		ChatId:       string(msg.ChatID),
		UserId:       string(msg.OwnerID),
		Text:         msg.Text,
		SentAtUnixMs: msg.CreatedAt.UnixMilli(),
	}
}

func newPbMessagesFromMessages(msgs []*models.Message) []*pb.Message {
	results := make([]*pb.Message, len(msgs))

	for i, msg := range msgs {
		results[i] = newPbMessageFromMessage(msg)
	}

	return results
}
