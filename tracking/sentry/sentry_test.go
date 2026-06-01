package sentry_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azcov/gokit/tracking"
	"github.com/azcov/gokit/tracking/sentry"
)

func TestNew_EmptyDSN(t *testing.T) {
	_, err := sentry.New(tracking.Config{})
	if err == nil {
		t.Error("expected error for empty DSN")
	}
}

func TestNew_ValidDSN(t *testing.T) {
	s, err := sentry.New(tracking.Config{DSN: "https://key@sentry.io/project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sentry")
	}
}

func TestCaptureError_Nil(t *testing.T) {
	s, _ := sentry.New(tracking.Config{DSN: "http://localhost"})
	err := s.CaptureError(context.Background(), nil)
	if err != nil {
		t.Errorf("CaptureError(nil) should return nil, got %v", err)
	}
}

func TestCaptureError_SendsToServer(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s, err := sentry.New(tracking.Config{
		DSN:         srv.URL,
		Environment: "test",
		Release:     "v1.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.CaptureError(context.Background(), errors.New("test error"),
		tracking.Tag{Key: "service", Value: "api"},
	); err != nil {
		t.Errorf("CaptureError: %v", err)
	}

	if !called {
		t.Error("expected Sentry server to be called")
	}
}

func TestCaptureMessage_SendsToServer(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s, _ := sentry.New(tracking.Config{DSN: srv.URL})
	if err := s.CaptureMessage(context.Background(), "something happened", tracking.LevelWarning,
		tracking.Tag{Key: "component", Value: "worker"},
	); err != nil {
		t.Errorf("CaptureMessage: %v", err)
	}

	if !called {
		t.Error("expected server to be called")
	}
}

func TestCapture_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	s, _ := sentry.New(tracking.Config{DSN: srv.URL})
	err := s.CaptureMessage(context.Background(), "msg", tracking.LevelError)
	if err == nil {
		t.Error("expected error on 400 response")
	}
}

func TestClose(t *testing.T) {
	s, _ := sentry.New(tracking.Config{DSN: "http://localhost"})
	if err := s.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestTrackerInterface(t *testing.T) {
	s, _ := sentry.New(tracking.Config{DSN: "http://localhost"})
	var _ tracking.Tracker = s
}
