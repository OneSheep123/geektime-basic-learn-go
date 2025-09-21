package job

import (
	"context"
	"ddd_demo/internal/domain"
	"ddd_demo/internal/service"
	"ddd_demo/pkg/logger"
	"fmt"
	"golang.org/x/sync/semaphore"
	"time"
)

// Executor 执行器
type Executor interface {
	Name() string
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	// 具体实现来控制
	// 真正去执行一个任务
	Exec(ctx context.Context, j domain.Job) error
}

// LocalFuncExecutor 本地的执行器
type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
	// fn func(ctx context.Context, j domain.Job)
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]func(ctx context.Context, j domain.Job) error)}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了？ %s", j.Name)
	}
	return fn(ctx, j)
}

// RegisterFunc 注册一个本地的任务
func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

// Scheduler 调度器
type Scheduler struct {
	svc     service.JobService
	l       logger.LoggerV1
	execs   map[string]Executor
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{svc: svc, l: l,
		limiter: semaphore.NewWeighted(200), // 本地执行器的并发限制
		execs:   make(map[string]Executor),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 尝试获取信号量
		// 这一操作本身是 同步阻塞 的，因为它会等待直到获取到信号量或者 ctx 被取消/超时
		// 注意：这里的acquire应该放在这里，不应该放在101-103中间，否则会出现preempt之后，又因为acquire问题，一直没有被执行
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		// 一次调度的数据库查询时间
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 你不能 return
			// 你要继续下一轮
			s.l.Error("抢占任务失败", logger.Error(err))
		}

		// 判断当前任务对应执行器是否存在
		exec, ok := s.execs[j.Executor]
		if !ok {
			// DEBUG 的时候最好中断
			// 线上就继续
			s.l.Error("未找到对应的执行器",
				logger.String("executor", j.Executor))
			continue
		}

		go func() {
			defer func() {
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					s.l.Error("释放任务失败",
						logger.Error(err1),
						logger.Int64("jid", j.Id))
				}
			}()
			// 异步执行，不要阻塞主调度循环
			// 执行完毕之后
			// 这边要考虑超时控制，任务的超时控制
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				// 你也可以考虑在这里重试
				s.l.Error("任务执行失败", logger.Error(err1))
			}
			// 你要不要考虑下一次调度？
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}
		}()
	}
}
