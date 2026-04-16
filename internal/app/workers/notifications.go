// Package workers provides background workers for notification processing.
package workers

import (
	"context"
)

type Consumer interface {
	Start(ctx context.Context) error
}

// Worker starts notification consumption jobs.
type Worker struct {
	consumer Consumer
}

// New creates a new Worker.
func New(consumer Consumer) *Worker {
	return &Worker{
		consumer: consumer,
	}
}

// Run starts the worker.
func (w *Worker) Run(ctx context.Context) error {
	return w.consumer.Start(ctx)
}
