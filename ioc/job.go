package ioc

import (
	"ddd_demo/internal/job"
	"ddd_demo/internal/service"
	"ddd_demo/pkg/logger"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankService, l logger.LoggerV1, client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, l, client, time.Second*30)
}

func InitJobs(l logger.LoggerV1, rjob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "定时任务执行",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", builder.Build(rjob))
	if err != nil {
		panic(err)
	}
	return expr
}
