package dto

type CreateProfileRequest struct {
	UserID    int64
	Nickname  string
	Bio       *string
	AvatarURL *string
}

type UpdateProfileRequest struct {
	UserID    int64
	Nickname  *string
	Bio       *string
	AvatarURL *string
}

type SearchByNicknameRequest struct {
	Query string
	Limit int64
}
