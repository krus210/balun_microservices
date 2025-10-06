package grpc

import (
	"social/internal/app/models"
	pb "social/pkg/api"
)

func newPbFriendRequestFromFriendRequest(fr *models.FriendRequest) *pb.FriendRequest {
	return &pb.FriendRequest{
		RequestId:  fr.ID,
		FromUserId: fr.FromUserID,
		ToUserId:   fr.ToUserID,
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

func newPbFriendRequestsIDsFromFriendRequests(frs []*models.FriendRequest) []int64 {
	results := make([]int64, len(frs), len(frs))

	for i, fr := range frs {
		results[i] = fr.ID
	}

	return results
}
