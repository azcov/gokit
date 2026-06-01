package slog

import (
	"context"
	"log/slog"

	"github.com/azcov/gokit/logger"
)

var _ logger.Logger = (*Slog)(nil)

type Slog struct {
	l *slog.Logger
}

func New(l *slog.Logger) *Slog {
	if l == nil {
		l = slog.Default()
	}
	return &Slog{l: l}
}

func (s *Slog) Debug(msg string, fields ...logger.Field) {
	s.l.Debug(msg, toArgs(fields)...)
}

func (s *Slog) Info(msg string, fields ...logger.Field) {
	s.l.Info(msg, toArgs(fields)...)
}

func (s *Slog) Warn(msg string, fields ...logger.Field) {
	s.l.Warn(msg, toArgs(fields)...)
}

func (s *Slog) Error(msg string, fields ...logger.Field) {
	s.l.Error(msg, toArgs(fields)...)
}

func (s *Slog) Fatal(msg string, fields ...logger.Field) {
	s.l.Error(msg, toArgs(fields)...)
}

func (s *Slog) DebugContext(ctx context.Context, msg string, fields ...logger.Field) {
	s.l.DebugContext(ctx, msg, toArgs(fields)...)
}

func (s *Slog) InfoContext(ctx context.Context, msg string, fields ...logger.Field) {
	s.l.InfoContext(ctx, msg, toArgs(fields)...)
}

func (s *Slog) WarnContext(ctx context.Context, msg string, fields ...logger.Field) {
	s.l.WarnContext(ctx, msg, toArgs(fields)...)
}

func (s *Slog) ErrorContext(ctx context.Context, msg string, fields ...logger.Field) {
	s.l.ErrorContext(ctx, msg, toArgs(fields)...)
}

func (s *Slog) FatalContext(ctx context.Context, msg string, fields ...logger.Field) {
	s.l.ErrorContext(ctx, msg, toArgs(fields)...)
}

func (s *Slog) With(fields ...logger.Field) logger.Logger {
	return &Slog{l: s.l.With(toArgs(fields)...)}
}

func toArgs(fields []logger.Field) []any {
	args := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}
	return args
}
