package chat_member

const chatMembersTable = "chat_members"

const (
	chatMembersTableColumnChatID = "chat_id"
	chatMembersTableColumnUserID = "user_id"
)

var chatMembersTableColumns = []string{
	chatMembersTableColumnChatID,
	chatMembersTableColumnUserID,
}
