package dto

type CreateProfileRequest struct {
	UserID    string
	Nickname  string
	Bio       *string
	AvatarURL *string
}

type UpdateProfileRequest struct {
	UserID    string
	Nickname  *string
	Bio       *string
	AvatarURL *string
}

type SearchByNicknameRequest struct {
	Query string
	Limit int64
}
