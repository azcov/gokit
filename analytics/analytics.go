package analytics

import "context"

type Properties map[string]any

type Event struct {
	Name       string
	UserID     string
	Properties Properties
}

type Trait struct {
	UserID     string
	Properties Properties
}

type PageView struct {
	UserID     string
	Name       string
	URL        string
	Properties Properties
}

type Config struct {
	APIKey   string `config:"api_key"`
	Endpoint string `config:"endpoint"`
}

type Tracker interface {
	Track(ctx context.Context, event Event) error
	Identify(ctx context.Context, trait Trait) error
	Page(ctx context.Context, page PageView) error
	Close() error
}
