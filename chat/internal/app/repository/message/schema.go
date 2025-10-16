package message

const MessagesTable = "messages"

const (
	MessagesTableColumnID        = "id"
	MessagesTableColumnText      = "text"
	MessagesTableColumnChatID    = "chat_id"
	MessagesTableColumnOwnerID   = "owner_id"
	MessagesTableColumnCreatedAt = "created_at"
	MessagesTableColumnUpdatedAt = "updated_at"
)

var MessagesTableColumns = []string{
	MessagesTableColumnID,
	MessagesTableColumnText,
	MessagesTableColumnChatID,
	MessagesTableColumnOwnerID,
	MessagesTableColumnCreatedAt,
	MessagesTableColumnUpdatedAt,
}
