package cron

import "context"

type Job struct {
	Name     string
	Schedule string
	Handler  func(ctx context.Context) error
}

type Scheduler interface {
	Add(job Job) error
	Remove(name string) error
	Start(ctx context.Context) error
	Stop() error
}
