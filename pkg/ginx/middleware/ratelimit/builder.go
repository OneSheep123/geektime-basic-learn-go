package ratelimit

import (
	"context"
	"ddd_demo/pkg/limiter"
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Builder struct {
	prefix  string
	limiter limiter.Limiter
}

func NewBuilder(l limiter.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: l,
	}
}

func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		if ctx.GetHeader("x-stress") == "true" {
			// 用 context.Context 来带这个标记位
			newCtx := context.WithValue(ctx, "x-stress", true)
			ctx.Request = ctx.Request.Clone(newCtx)
			ctx.Next()
			return
		}

		limited, err := b.limiter.Limit(ctx, fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP()))
		if err != nil {
			log.Println(err)
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			// 保守做法：因为借助于 Redis 来做限流，那么 Redis 崩溃了，为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// 激进做法：虽然 Redis 崩溃了，但是这个时候还是要尽量服务正常的用户，所以不限流
			// ctx.Next()
			return
		}
		if limited {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
