package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Id   int64
	Name string

	Executor string // 执行器，local 或者 remote

	Cron string
	Cfg  string

	CancelFunc func() error
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
	cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	// 你怎么算？要根据 cron 表达式来算
	// 可以做成包变量，因为基本不可能变

	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
