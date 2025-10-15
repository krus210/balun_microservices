package message

const messagesTable = "messages"

const (
	messagesTableColumnID        = "id"
	messagesTableColumnText      = "text"
	messagesTableColumnChatID    = "chat_id"
	messagesTableColumnOwnerID   = "owner_id"
	messagesTableColumnCreatedAt = "created_at"
	messagesTableColumnUpdatedAt = "updated_at"
)

var messagesTableColumns = []string{
	messagesTableColumnID,
	messagesTableColumnText,
	messagesTableColumnChatID,
	messagesTableColumnOwnerID,
	messagesTableColumnCreatedAt,
	messagesTableColumnUpdatedAt,
}
