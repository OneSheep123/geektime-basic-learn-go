package circuitbreaker

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InterceptorBuilder 构建熔断拦截器的结构体
// 基于 go-kratos/aegis 的熔断器实现，用于 gRPC 服务端拦截器
type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		err = b.breaker.Allow()
		if err == nil {
			resp, err = handler(ctx, req)
			if err == nil {
				b.breaker.MarkSuccess()
			} else {
				// 更加仔细检测，只有真实代表服务端出现故障的，才 mark failed
				b.breaker.MarkFailed()
			}
			return
		} else {
			b.breaker.MarkFailed()
			return nil, status.Errorf(codes.Unavailable, "熔断")
		}
	}
}
