package chat

const ChatsTable = "chats"

const (
	ChatsTableColumnID        = "id"
	ChatsTableColumnCreatedAt = "created_at"
	ChatsTableColumnUpdatedAt = "updated_at"
)

var ChatsTableColumns = []string{
	ChatsTableColumnID,
	ChatsTableColumnCreatedAt,
	ChatsTableColumnUpdatedAt,
}
