package grpc

import (
	"context"

	pb "auth/pkg/api"
)

func (h *AuthController) GetJWKS(ctx context.Context, req *pb.GetJWKSRequest) (*pb.GetJWKSResponse, error) {
	jwksResponse, err := h.usecase.GetJWKS(ctx)
	if err != nil {
		return nil, err
	}

	// Конвертируем DTO в protobuf
	pbJWKs := make([]*pb.JWK, 0, len(jwksResponse.Keys))
	for _, key := range jwksResponse.Keys {
		pbJWKs = append(pbJWKs, &pb.JWK{
			Kty: key.KTY,
			Use: key.Use,
			Kid: key.KID,
			Alg: key.Alg,
			N:   key.N,
			E:   key.E,
		})
	}

	return &pb.GetJWKSResponse{
		Jwks: pbJWKs,
	}, nil
}
