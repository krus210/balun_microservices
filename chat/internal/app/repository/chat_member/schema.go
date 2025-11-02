package chat_member

const ChatMembersTable = "chat_members"

const (
	ChatMembersTableColumnChatID = "chat_id"
	ChatMembersTableColumnUserID = "user_id"
)

var ChatMembersTableColumns = []string{
	ChatMembersTableColumnChatID,
	ChatMembersTableColumnUserID,
}
