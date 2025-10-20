package grpc

import (
	pb "github.com/aarondever/notiflow/proto/email"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(
	NewEmailGRPCHandler,
)

// RegisterEmailService registers the email gRPC service
func RegisterEmailService(s *grpc.Server, handler *EmailGRPCHandler) {
	pb.RegisterEmailServiceServer(s, handler)
}
