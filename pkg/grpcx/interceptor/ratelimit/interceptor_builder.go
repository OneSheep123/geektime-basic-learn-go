package ratelimit

import (
	"context"
	"ddd_demo/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

// NewInterceptorBuilder key 1. limiter:interactive-service => 整个点赞的应用限流
func NewInterceptorBuilder(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 你有保守的做法，也有激进的做法
			// 这个是保守的做法
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 检查是否触发限流
		limited, err := b.limiter.Limit(ctx, b.key)
		// 如果限流器出错或者触发了限流，则在 Context 中设置降级标记
		// 后续的业务处理逻辑可以通过 ctx.Value("downgrade") 检测到该标记，
		// 并执行相应的降级策略（如返回缓存数据、简化响应、返回默认值等）
		if err != nil || limited {
			ctx = context.WithValue(ctx, "downgrade", "true")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			// 这个 key，limiter:UserService
			limited, err := b.limiter.Limit(ctx, b.key)
			if err != nil {
				// 你有保守的做法，也有激进的做法
				// 这个是保守的做法
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}

			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "限流")
			}
		}
		return handler(ctx, req)
	}
}
