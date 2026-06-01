package tracking

import "context"

type Level string

const (
	LevelDebug   Level = "debug"
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
	LevelFatal   Level = "fatal"
)

type Tag struct {
	Key   string
	Value string
}

type Config struct {
	DSN         string `config:"dsn"`
	Environment string `config:"environment"`
	Release     string `config:"release"`
}

type Tracker interface {
	CaptureError(ctx context.Context, err error, tags ...Tag) error
	CaptureMessage(ctx context.Context, msg string, level Level, tags ...Tag) error
	Close() error
}
