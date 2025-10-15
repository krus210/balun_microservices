package chat

const chatsTable = "chats"

const (
	chatsTableColumnID        = "id"
	chatsTableColumnCreatedAt = "created_at"
	chatsTableColumnUpdatedAt = "updated_at"
)

var chatsTableColumns = []string{
	chatsTableColumnID,
	chatsTableColumnCreatedAt,
	chatsTableColumnUpdatedAt,
}
