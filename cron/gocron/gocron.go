package gocron

import (
	"context"

	gokitcron "github.com/azcov/gokit/cron"
	gcron "github.com/go-co-op/gocron/v2"
)

var _ gokitcron.Scheduler = (*GoCron)(nil)

type GoCron struct {
	s gcron.Scheduler
}

func New() (*GoCron, error) {
	s, err := gcron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &GoCron{s: s}, nil
}

func (g *GoCron) Add(job gokitcron.Job) error {
	_, err := g.s.NewJob(
		gcron.CronJob(job.Schedule, false),
		gcron.NewTask(func() { _ = job.Handler(context.Background()) }),
		gcron.WithName(job.Name),
	)
	return err
}

func (g *GoCron) Remove(name string) error {
	for _, j := range g.s.Jobs() {
		if j.Name() == name {
			return g.s.RemoveJob(j.ID())
		}
	}
	return nil
}

func (g *GoCron) Start(ctx context.Context) error {
	g.s.Start()
	<-ctx.Done()
	return g.s.Shutdown()
}

func (g *GoCron) Stop() error {
	return g.s.Shutdown()
}
