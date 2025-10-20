package friend_request

const FriendRequestsTable = "friend_requests"

const (
	FriendRequestsTableColumnID         = "id"
	FriendRequestsTableColumnFromUserID = "from_user_id"
	FriendRequestsTableColumnToUserID   = "to_user_id"
	FriendRequestsTableColumnStatus     = "status"
	FriendRequestsTableColumnCreatedAt  = "created_at"
	FriendRequestsTableColumnUpdatedAt  = "updated_at"
)

var FriendRequestsTableColumns = []string{
	FriendRequestsTableColumnID,
	FriendRequestsTableColumnFromUserID,
	FriendRequestsTableColumnToUserID,
	FriendRequestsTableColumnStatus,
	FriendRequestsTableColumnCreatedAt,
	FriendRequestsTableColumnUpdatedAt,
}
