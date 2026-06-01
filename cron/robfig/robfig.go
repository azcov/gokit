package robfig

import (
	"context"
	"sync"

	gokitcron "github.com/azcov/gokit/cron"
	robfigcron "github.com/robfig/cron/v3"
)

var _ gokitcron.Scheduler = (*Robfig)(nil)

type Robfig struct {
	c   *robfigcron.Cron
	mu  sync.Mutex
	ids map[string]robfigcron.EntryID
}

func New() *Robfig {
	return &Robfig{
		c:   robfigcron.New(robfigcron.WithSeconds()),
		ids: make(map[string]robfigcron.EntryID),
	}
}

func (r *Robfig) Add(job gokitcron.Job) error {
	id, err := r.c.AddFunc(job.Schedule, func() {
		_ = job.Handler(context.Background())
	})
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.ids[job.Name] = id
	r.mu.Unlock()
	return nil
}

func (r *Robfig) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if id, ok := r.ids[name]; ok {
		r.c.Remove(id)
		delete(r.ids, name)
	}
	return nil
}

func (r *Robfig) Start(ctx context.Context) error {
	r.c.Start()
	<-ctx.Done()
	r.c.Stop()
	return nil
}

func (r *Robfig) Stop() error {
	r.c.Stop()
	return nil
}
