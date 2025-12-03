package pool

import (
	"context"
)

type WorkerEvent uint

const (
	EventAllTaskDone = iota
)

type TaskFunction[I interface{}, O interface{}] func(context context.Context, input I) O
type Task[I interface{}, O interface{}] struct {
	TaskFunction TaskFunction[I, O]
	Input        I
}

type WorkerPool[I interface{}, O interface{}] interface {
	Submit(task Task[I, O]) error
	Start()
	Stop()
	Cancel()
	CloseSubmit()
	GetOutputChannel() chan O
	GetEventChannel() chan WorkerEvent
	GetTaskCount() int
}
