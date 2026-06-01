// Package queue defines background-job interfaces (Enqueuer, Worker). Each task
// is processed by exactly one competing consumer, with retry and scheduling.
//
// Implementations: queue/asynq.
package queue
