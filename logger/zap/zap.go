package zap

import (
	"context"

	"github.com/azcov/gokit/logger"
	"go.uber.org/zap"
)

var _ logger.Logger = (*Zap)(nil)

type Zap struct {
	z *zap.Logger
}

func New(z *zap.Logger) *Zap {
	return &Zap{z: z}
}

func NewProduction() (*Zap, error) {
	z, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Zap{z: z}, nil
}

func NewDevelopment() (*Zap, error) {
	z, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &Zap{z: z}, nil
}

func (l *Zap) Debug(msg string, fields ...logger.Field) {
	l.z.Debug(msg, toZap(fields)...)
}

func (l *Zap) Info(msg string, fields ...logger.Field) {
	l.z.Info(msg, toZap(fields)...)
}

func (l *Zap) Warn(msg string, fields ...logger.Field) {
	l.z.Warn(msg, toZap(fields)...)
}

func (l *Zap) Error(msg string, fields ...logger.Field) {
	l.z.Error(msg, toZap(fields)...)
}

func (l *Zap) Fatal(msg string, fields ...logger.Field) {
	l.z.Fatal(msg, toZap(fields)...)
}

func (l *Zap) DebugContext(_ context.Context, msg string, fields ...logger.Field) {
	l.z.Debug(msg, toZap(fields)...)
}

func (l *Zap) InfoContext(_ context.Context, msg string, fields ...logger.Field) {
	l.z.Info(msg, toZap(fields)...)
}

func (l *Zap) WarnContext(_ context.Context, msg string, fields ...logger.Field) {
	l.z.Warn(msg, toZap(fields)...)
}

func (l *Zap) ErrorContext(_ context.Context, msg string, fields ...logger.Field) {
	l.z.Error(msg, toZap(fields)...)
}

func (l *Zap) FatalContext(_ context.Context, msg string, fields ...logger.Field) {
	l.z.Fatal(msg, toZap(fields)...)
}

func (l *Zap) With(fields ...logger.Field) logger.Logger {
	return &Zap{z: l.z.With(toZap(fields)...)}
}

func toZap(fields []logger.Field) []zap.Field {
	out := make([]zap.Field, len(fields))
	for i, f := range fields {
		out[i] = zap.Any(f.Key, f.Value)
	}
	return out
}
