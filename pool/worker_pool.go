// Package pool provides a generic worker pool implementation for concurrent
// task processing.
//
// The pool uses Go generics to support type-safe task input and output.
// Workers process tasks from a shared queue and send results to an output channel.
// The pool supports graceful shutdown, task cancellation, and completion events.
//
// Basic usage:
//
//	pool := NewDefaultWorkerPool[InputType, OutputType](numWorkers)
//	pool.Start()
//	defer pool.Stop()
//
//	// Submit tasks
//	pool.Submit(Task{TaskFunction: myFunc, Input: input})
//
//	// Read results
//	result := <-pool.GetOutputChannel()
package pool

import (
	"context"
)

// WorkerEvent represents events emitted by the worker pool.
type WorkerEvent uint

const (
	// EventAllTaskDone is emitted when all submitted tasks have been processed
	// and CloseSubmit has been called.
	EventAllTaskDone = iota
)

// TaskFunction is the signature for task processing functions.
// It receives a context for cancellation and the task input,
// returning the processed output.
type TaskFunction[I interface{}, O interface{}] func(context context.Context, input I) O

// Task represents a unit of work to be processed by the worker pool.
type Task[I interface{}, O interface{}] struct {
	// TaskFunction is the function to execute for this task.
	TaskFunction TaskFunction[I, O]
	// Input is the data to pass to the task function.
	Input I
}

// WorkerPool defines the interface for a generic worker pool.
//
// Type parameters:
//   - I: the input type for tasks
//   - O: the output type for tasks
type WorkerPool[I interface{}, O interface{}] interface {
	// Submit adds a task to the pool's queue for processing.
	// Returns an error if the pool is not running.
	Submit(task Task[I, O]) error

	// Start initializes and starts the worker goroutines.
	Start()

	// Stop cancels all pending tasks and shuts down the workers.
	Stop()

	// Cancel signals all running tasks to stop via context cancellation.
	Cancel()

	// CloseSubmit indicates that no more tasks will be submitted.
	// EventAllTaskDone will be emitted once all current tasks complete.
	CloseSubmit()

	// GetOutputChannel returns the channel where task results are sent.
	GetOutputChannel() chan O

	// GetEventChannel returns the channel for worker pool events.
	GetEventChannel() chan WorkerEvent

	// GetTaskCount returns the number of tasks currently in the queue or processing.
	GetTaskCount() int
}
