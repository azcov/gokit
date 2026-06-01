package logger

import "context"

type Field struct {
	Key   string
	Value any
}

func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	DebugContext(ctx context.Context, msg string, fields ...Field)
	InfoContext(ctx context.Context, msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)
	ErrorContext(ctx context.Context, msg string, fields ...Field)
	FatalContext(ctx context.Context, msg string, fields ...Field)

	With(fields ...Field) Logger
}

// Nop is a no-op logger useful in tests and as a safe default.
type Nop struct{}

var _ Logger = (*Nop)(nil)

func (Nop) Debug(_ string, _ ...Field)                           {}
func (Nop) Info(_ string, _ ...Field)                            {}
func (Nop) Warn(_ string, _ ...Field)                            {}
func (Nop) Error(_ string, _ ...Field)                           {}
func (Nop) Fatal(_ string, _ ...Field)                           {}
func (Nop) DebugContext(_ context.Context, _ string, _ ...Field) {}
func (Nop) InfoContext(_ context.Context, _ string, _ ...Field)  {}
func (Nop) WarnContext(_ context.Context, _ string, _ ...Field)  {}
func (Nop) ErrorContext(_ context.Context, _ string, _ ...Field) {}
func (Nop) FatalContext(_ context.Context, _ string, _ ...Field) {}
func (Nop) With(_ ...Field) Logger                               { return Nop{} }
