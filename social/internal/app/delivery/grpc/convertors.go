package grpc

import (
	"social/internal/app/models"
	pb "social/pkg/api"
)

func newPbFriendRequestFromFriendRequest(fr *models.FriendRequest) *pb.FriendRequest {
	return &pb.FriendRequest{
		RequestId:  string(fr.ID),
		FromUserId: string(fr.FromUserID),
		ToUserId:   string(fr.ToUserID),
		Status:     pb.FriendRequestStatus(fr.Status),
	}
}

func newPbFriendRequestsFromFriendRequests(frs []*models.FriendRequest) []*pb.FriendRequest {
	results := make([]*pb.FriendRequest, len(frs), len(frs))

	for i, fr := range frs {
		results[i] = newPbFriendRequestFromFriendRequest(fr)
	}

	return results
}

func newPbFriendRequestsIDsFromFriendRequests(frs []*models.FriendRequest) []string {
	results := make([]string, len(frs), len(frs))

	for i, fr := range frs {
		results[i] = string(fr.ID)
	}

	return results
}
