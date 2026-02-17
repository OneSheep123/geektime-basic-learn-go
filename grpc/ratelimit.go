package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

// CounterLimiter 基于计数器的限流器
// 使用原子操作维护当前并发请求数，当并发数超过阈值时拒绝请求
type CounterLimiter struct {
	cnt       atomic.Int32 // 当前并发请求数
	threshold int32        // 并发请求阈值
}

// BuildServerInterceptor 构建gRPC服务端拦截器
// 每个请求会原子递增计数器，处理完成后递减
// 如果当前并发数超过阈值，返回 ResourceExhausted 错误
func (c *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		// 请求进来，先占坑
		cnt := c.cnt.Add(1)
		defer func() {
			c.cnt.Add(-1)
		}()
		if cnt <= c.threshold {
			resp, err = handler(ctx, req)
			// 返回了响应
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// FixedWindowLimiter 固定窗口限流器
// 将时间划分为固定大小的窗口，每个窗口内允许的最大请求数固定
// 窗口结束后计数器重置
type FixedWindowLimiter struct {
	window          time.Duration // 窗口大小
	lastWindowStart time.Time     // 当前窗口开始时间
	cnt             int           // 当前窗口请求数
	threshold       int           // 窗口内最大请求数
	lock            sync.Mutex    // 保护计数器和窗口状态的锁
}

// NewFixedWindowLimiter 创建一个新的固定窗口限流器
// window: 窗口大小，threshold: 每个窗口内允许的最大请求数
func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{window: window, lastWindowStart: time.Now(), cnt: 0, threshold: threshold}

}

// BuildServerInterceptor 构建gRPC服务端拦截器
// 检查当前窗口是否过期，过期则重置计数器
// 如果当前窗口内请求数超过阈值，拒绝请求
func (c *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		c.lock.Lock()
		now := time.Now()
		if now.After(c.lastWindowStart.Add(c.window)) {
			c.cnt = 0
			c.lastWindowStart = now
		}
		cnt := c.cnt + 1
		c.lock.Unlock()
		if cnt <= c.threshold {
			resp, err = handler(ctx, req)
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// SlidingWindowLimiter 滑动窗口限流器
// 使用优先级队列记录每个请求的时间戳
// 只统计窗口时间内的请求，比固定窗口更精确
type SlidingWindowLimiter struct {
	window time.Duration // 窗口大小
	// 请求到来的时间戳
	// 时间戳最小的在队首
	queue     queue.PriorityQueue[time.Time] // 请求时间戳队列
	lock      sync.Mutex                     // 保护队列的锁
	threshold int                            // 窗口内最大请求数
}

// BuildServerInterceptor 构建gRPC服务端拦截器
// 清理窗口外的时间戳，统计窗口内的请求数
// 如果窗口内请求数超过阈值，拒绝请求
func (c *SlidingWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		c.lock.Lock()
		// 我先考虑队列里面的时间戳是不是都在我的窗口范围内
		now := time.Now()

		// 快路径检测
		if c.queue.Len() < c.threshold {
			_ = c.queue.Enqueue(now)
			c.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}

		windowStart := now.Add(-c.window)
		for {
			first, _ := c.queue.Peek()
			if first.Before(windowStart) {
				// 把第一个元素删了
				_, _ = c.queue.Dequeue()
			} else {
				break
			}
		}
		if c.queue.Len() < c.threshold {
			_ = c.queue.Enqueue(now)
			c.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}
		c.lock.Unlock()
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// TokenBucketLimiter 令牌桶限流器
// 以固定速率生成令牌，请求需要获取令牌才能执行
// 桶有容量限制，多余的令牌会被丢弃
type TokenBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration // 令牌生成间隔
	buckets   chan struct{} // 令牌桶（使用chan实现）
	closeCh   chan struct{} // 关闭信号
	closeOnce sync.Once     // 确保只关闭一次
}

// BuildServerInterceptor 构建gRPC服务端拦截器
// 启动定时器生成令牌
// 每个请求尝试从桶中获取令牌，获取成功则执行，否则拒绝
func (c *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				select {
				case c.buckets <- struct{}{}:
				default:
					// bucket 满了
				}
			case <-c.closeCh:
				return
			}
		}
	}()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-c.buckets:
			return handler(ctx, req)
		//做法1
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
			// 做法2
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}

// Close 关闭限流器，释放资源
func (c *TokenBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}

// LeakyBucketLimiter 漏桶限流器
// 以固定速率处理请求，请求先进入桶中，然后以固定速率流出处理
// 桶有容量限制，满了之后新请求会被拒绝
type LeakyBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration // 请求处理间隔（漏水速率）
	closeCh   chan struct{} // 关闭信号
	closeOnce sync.Once     // 确保只关闭一次
}

// BuildServerInterceptor 构建gRPC服务端拦截器
// 每个请求需要等待定时器触发才能执行，实现匀速处理
// 如果限流器已关闭，拒绝请求
func (c *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(c.interval)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ticker.C:
			return handler(ctx, req)
		case <-c.closeCh:
			// 限流器已经关了
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		//做法1
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
			// 做法2
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}

// Close 关闭限流器，释放资源
func (c *LeakyBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}

//1. CounterLimiter（计数器限流器）
//- 结构体说明：基于原子操作维护并发请求数
//- 方法说明：请求进入时占坑计数，处理完成递减，超过阈值则限流
//
//2. FixedWindowLimiter（固定窗口限流器）
//- 结构体说明：时间划分为固定窗口，每个窗口有独立的请求计数
//- 构造函数 NewFixedWindowLimiter：创建指定窗口大小和阈值的限流器
//- 方法说明：窗口过期重置计数器，超过阈值拒绝请求
//
//3. SlidingWindowLimiter（滑动窗口限流器）
//- 结构体说明：使用优先级队列记录请求时间戳，更精确的窗口统计
//- 方法说明：清理窗口外的时间戳，快路径优化，超过阈值限流
//
//4. TokenBucketLimiter（令牌桶限流器）
//- 结构体说明：固定速率生成令牌，请求需获取令牌才能执行
//- 方法说明：启动定时器生成令牌，获取到令牌则执行，否则拒绝
//- Close 方法：关闭限流器释放资源
//
//5. LeakyBucketLimiter（漏桶限流器）
//- 结构体说明：固定速率处理请求，请求需等待定时器触发
//- 方法说明：匀速处理请求，限流器关闭后拒绝新请求
//- Close 方法：关闭限流器释放资源
