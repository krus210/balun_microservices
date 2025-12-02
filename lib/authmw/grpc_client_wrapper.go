package authmw

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// GRPCClientWrapper оборачивает gRPC соединение для вызова GetJWKS
type GRPCClientWrapper struct {
	conn        *grpc.ClientConn
	serviceName string
	methodName  string
}

// NewGRPCClientWrapper создаёт wrapper для вызова GetJWKS через generic grpc
func NewGRPCClientWrapper(conn *grpc.ClientConn) *GRPCClientWrapper {
	return &GRPCClientWrapper{
		conn:        conn,
		serviceName: "github.com.krus210.balun_microservices.protobuf.auth.v1.proto.AuthService",
		methodName:  "GetJWKS",
	}
}

// GetJWKS вызывает метод GetJWKS через generic grpc.Invoke
// Использует protobuf типы из jwks.pb.go
func (w *GRPCClientWrapper) GetJWKS(ctx context.Context) (*GetJWKSResponse, error) {
	// Формируем полный путь к методу
	method := fmt.Sprintf("/%s/%s", w.serviceName, w.methodName)

	// Создаём пустой protobuf запрос
	req := &GetJWKSRequest{}

	// Создаём protobuf ответ
	resp := &GetJWKSResponse{}

	// Вызываем метод через generic invoke
	err := w.conn.Invoke(ctx, method, req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke GetJWKS: %w", err)
	}

	return resp, nil
}

// Убеждаемся что GRPCClientWrapper реализует AuthServiceClient
var _ AuthServiceClient = (*GRPCClientWrapper)(nil)
