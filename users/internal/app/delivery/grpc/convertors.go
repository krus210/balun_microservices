package grpc

import (
	"users/internal/app/models"
	pb "users/pkg/api"
)

func newPbUserProfileFromUserProfile(userProfile *models.UserProfile) *pb.UserProfile {
	return &pb.UserProfile{
		UserId:    userProfile.UserID,
		Nickname:  userProfile.Nickname,
		Bio:       userProfile.Bio,
		AvatarUrl: userProfile.AvatarURL,
	}
}

func newPbUserProfilesFromUserProfiles(userProfiles []*models.UserProfile) []*pb.UserProfile {
	results := make([]*pb.UserProfile, len(userProfiles), len(userProfiles))

	for i, userProfile := range userProfiles {
		results[i] = newPbUserProfileFromUserProfile(userProfile)
	}

	return results
}
