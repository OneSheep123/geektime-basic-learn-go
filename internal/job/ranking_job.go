package job

import (
	"context"
	"ddd_demo/internal/service"
	"time"
)

type RankingJob struct {
	svc     service.RankService
	timeout time.Duration
}

func NewRankingJob(svc service.RankService, timeout time.Duration) *RankingJob {
	return &RankingJob{svc: svc, timeout: timeout}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return r.svc.TopN(ctx)
}
