package user

const UserProfilesTable = "user_profiles"

const (
	UserProfilesTableColumnID        = "id"
	UserProfilesTableColumnNickname  = "nickname"
	UserProfilesTableColumnBio       = "bio"
	UserProfilesTableColumnAvatarURL = "avatar_url"
	UserProfilesTableColumnCreatedAt = "created_at"
	UserProfilesTableColumnUpdatedAt = "updated_at"
)

var UserProfilesTableColumns = []string{
	UserProfilesTableColumnID,
	UserProfilesTableColumnNickname,
	UserProfilesTableColumnBio,
	UserProfilesTableColumnAvatarURL,
	UserProfilesTableColumnCreatedAt,
	UserProfilesTableColumnUpdatedAt,
}
