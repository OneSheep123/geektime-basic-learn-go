package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func (g *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := g.db.WithContext(ctx)
	for {
		now := time.Now().UnixMilli()
		var j Job
		// 分布式任务调度系统
		// 防止拿不到资源，不断循环执行db操作
		// 1. 一次拉一批，我一次性取出 100 条来，然后，我随机从某一条开始，向后开始抢占
		// 2. 我搞个随机偏移量，0-100 生成一个随机偏移量。兜底：第一轮没查到，偏移量回归到 0
		// 3. 我搞一个 id 取余分配，status = ? AND next_time <=? AND id%10 = ? 兜底：不加余数条件，取next_time 最老的
		// 查询条件：
		// 1. 等待调度的任务：status = waiting AND next_time <= now
		// 2. 续约失败的任务：status = running AND utime < (now - 3分钟)，表示曾经有人调度但续约失败
		ddl := now - (time.Minute * 3).Milliseconds()
		err := db.Where("(status = ? AND next_time <?) OR (status = ? AND utime < ?)",
			jobStatusWaiting, now, jobStatusRunning, ddl).Error
		// 你找到了，可以被抢占的
		// 找到之后你要干嘛？你要抢占
		if err != nil {
			// // 没有任务。从这里返回
			return Job{}, err
		}
		// 两个 goroutine 都拿到 id =1 的数据
		// 能不能用 utime?
		// 乐观锁，CAS 操作，compare AND Swap
		// 有一个很常见的面试刷亮点：就是用乐观锁取代 FOR UPDATE
		// 面试套路（性能优化）：曾将用了 FOR UPDATE =>性能差，还会有死锁 => 我优化成了乐观锁
		res := db.Where("id=? AND version = ?",
			j.Id, j.Version).Model(&Job{}).
			Updates(map[string]any{
				"status":  jobStatusRunning,
				"utime":   now,
				"version": j.Version + 1,
			})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			// 抢占失败，你只能说，我要继续下一轮
			continue
		}
		return j, nil
	}
}

func (g *GORMJobDAO) Release(ctx context.Context, id int64) error {
	// 这里有一个问题。你要不要检测 status 或者 version?
	// WHERE version = ?
	// (但这种场景比较少，一般就是节点运行中，可能连不上数据库了，然后数据刚好被其他节点强占执行中；但是一旦连不上数据库，其他节点都也连不上的)
	// 要。你们的作业记得修改
	return g.db.WithContext(ctx).Model(&Job{}).Where("id =?", id).
		Updates(map[string]any{
			"status": jobStatusWaiting,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

func (g *GORMJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id =?", id).
		Updates(map[string]any{
			"utime": time.Now().UnixMilli(),
		}).Error
}

func (g *GORMJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id =?", id).
		Updates(map[string]any{
			"next_time": next.UnixMilli(),
			"utime":     time.Now().UnixMilli(),
		}).Error
}

func (g *GORMJobDAO) Stop(ctx context.Context, id int64) error {
	return g.db.WithContext(ctx).Model(&Job{}).Where("id =?", id).
		Updates(map[string]any{
			"status": jobStatusPaused,
			"utime":  time.Now().UnixMilli(),
		}).Error
}

type Job struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	Config string `gorm:"type:varchar(255);not null;default:'';comment:'配置'"`
	Name   string `gorm:"unique"` // 唯一索引，任务名称

	Executor string // 执行器，local 或者 remote
	Status   int

	NextTime int64 `gorm:"index"`

	// cron 表达式
	Cron string // 执行频率

	Version int

	Ctime int64
	Utime int64
}

const (
	jobStatusWaiting = iota
	// 已经被抢占
	jobStatusRunning
	// 还可以有别的取值

	// 暂停调度
	jobStatusPaused
)
