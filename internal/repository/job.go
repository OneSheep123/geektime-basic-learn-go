package repository

import (
	"context"
	"ddd_demo/internal/domain"
	"ddd_demo/internal/repository/dao"
	"time"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type PreemptCronJobRepository struct {
	dao dao.JobDAO
}

func (g *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := g.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Id:       j.Id,
		Name:     j.Name,
		Cfg:      j.Config,
		Cron:     j.Cron,
		Executor: j.Executor,
	}, nil
}

func (g *PreemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return g.dao.Release(ctx, id)
}

func (g *PreemptCronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return g.dao.UpdateUtime(ctx, id)
}

func (g *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.dao.UpdateNextTime(ctx, id, next)
}

func (g *PreemptCronJobRepository) Stop(ctx context.Context, id int64) error {
	return g.dao.Stop(ctx, id)
}
