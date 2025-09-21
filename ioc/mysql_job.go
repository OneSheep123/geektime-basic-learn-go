package ioc

import (
	"context"
	"ddd_demo/internal/domain"
	"ddd_demo/internal/job"
	"ddd_demo/internal/service"
	"ddd_demo/pkg/logger"
	"time"
)

// InitLocalFuncExecutor 初始化本地的执行器
func InitLocalFuncExecutor(svc service.RankService) *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	// 要在数据库里面插入一条记录。
	// ranking job 的记录，通过管理任务接口来插入
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}

func InitScheduler(l logger.LoggerV1,
	local *job.LocalFuncExecutor,
	svc service.JobService) *job.Scheduler {
	// 初始化调度器
	res := job.NewScheduler(svc, l)
	// 注册本地的执行器
	res.RegisterExecutor(local)
	// 注册远程的执行器
	// res.RegisterExecutor(remote)
	return res
}
