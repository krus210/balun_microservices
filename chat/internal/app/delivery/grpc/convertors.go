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

	return &pb.Chat{
		ChatId:            chat.ID,
		ParticipantIds:    chat.ParticipantIDs,
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
		MessageId:    msg.ID,
		ChatId:       msg.ChatID,
		UserId:       msg.OwnerID,
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
