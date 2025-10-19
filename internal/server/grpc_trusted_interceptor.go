package server

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TrustedSubnetInterceptor(allowedSubnet string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if allowedSubnet == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("No metadata in request")
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		realIPHeaders := md.Get("x-real-ip")
		if len(realIPHeaders) == 0 {
			logger.Warn("X-Real-IP header is missing")
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		realIP := realIPHeaders[0]
		ip := net.ParseIP(realIP)
		if ip == nil {
			logger.Warn("Invalid IP address in X-Real-IP header", zap.String("ip", realIP))
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		_, subnet, err := net.ParseCIDR(allowedSubnet)
		if err != nil {
			logger.Error("Invalid subnet configuration",
				zap.String("subnet", allowedSubnet),
				zap.Error(err))
			return nil, status.Error(codes.Internal, "internal server error")
		}

		if !subnet.Contains(ip) {
			logger.Warn("IP address not in allowed subnet",
				zap.String("ip", realIP),
				zap.String("subnet", allowedSubnet))
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		return handler(ctx, req)
	}
}
