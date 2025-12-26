package finder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"fdups/hasher"
	"fdups/log"
	"fdups/pool"

	"go.uber.org/zap"
)

// taskInput is the input type for hash computation tasks.
type taskInput struct {
	fileInfo *FileInfo
}

// taskOutput is the output type for hash computation tasks.
type taskOutput struct {
	fileInfo *FileInfo
	err      error
}

// walkDirectoryYield represents a result from directory traversal.
type walkDirectoryYield struct {
	err      error
	fileInfo *FileInfo
}

// FileFilter is a predicate function that determines if a file should be processed.
// It receives the file path and os.FileInfo, returning true if the file should be included.
type FileFilter func(path string, info os.FileInfo) bool

// baseFinder provides the common implementation for all Finder types.
// It handles directory traversal, worker pool management, and result aggregation.
// Concrete finders embed baseFinder and configure it with specific hashers and filters.
type baseFinder struct {
	targetDirectory string
	workerPool      pool.WorkerPool[taskInput, taskOutput]
	result          map[string][]FileInfo
	hasher          hasher.Hasher
	fileFilter      FileFilter
}

// newBaseFinder creates a new baseFinder with the specified configuration.
// The worker pool is sized to the number of available CPU cores.
func newBaseFinder(targetDirectory string, h hasher.Hasher, filter FileFilter) *baseFinder {
	return &baseFinder{
		targetDirectory: targetDirectory,
		workerPool:      pool.NewDefaultWorkerPool[taskInput, taskOutput](runtime.NumCPU()),
		result:          make(map[string][]FileInfo),
		hasher:          h,
		fileFilter:      filter,
	}
}

func (f *baseFinder) Find() (error, map[string][]FileInfo) {
	f.workerPool.Start()
	log.L().Debug("Worker pool started")
	defer f.stopWorkerPool()

	errorChannel := make(chan error)
	log.L().Debug("Error handling channel created")
	defer close(errorChannel)

	go f.runWalkGoroutine(errorChannel)
	go f.runCollectGoroutine(errorChannel)

	return f.waitForCompletion(errorChannel)
}

func (f *baseFinder) stopWorkerPool() {
	f.workerPool.Stop()
	log.L().Debug("Worker pool stopped")
}

func (f *baseFinder) runWalkGoroutine(errorChannel chan<- error) {
	for item := range f.walkDirectory() {
		if item.err != nil {
			log.L().Error("Error received, aborting", zap.Error(item.err))
			errorChannel <- item.err
			return
		}
		f.submitHashTask(item.fileInfo)
	}
	f.workerPool.CloseSubmit()
	log.L().Debug("All task submitted; Goroutine exit")
	errorChannel <- nil
}

func (f *baseFinder) submitHashTask(fileInfo *FileInfo) {
	_ = f.workerPool.Submit(pool.Task[taskInput, taskOutput]{
		TaskFunction: f.createHashFunction(),
		Input:        taskInput{fileInfo: fileInfo},
	})
	log.L().Debug("Hashing task submitted", zap.String("name", fileInfo.Name))
}

func (f *baseFinder) runCollectGoroutine(errorChannel chan<- error) {
	for {
		select {
		case item := <-f.workerPool.GetOutputChannel():
			if item.err != nil {
				log.L().Error("Error received, aborting", zap.Error(item.err))
				errorChannel <- item.err
				return
			}
			f.groupDuplicates(*item.fileInfo)
		case event := <-f.workerPool.GetEventChannel():
			if event == pool.EventAllTaskDone {
				log.L().Debug("All task processed; Goroutine exit")
				errorChannel <- nil
				return
			}
		}
	}
}

func (f *baseFinder) waitForCompletion(errorChannel <-chan error) (error, map[string][]FileInfo) {
	for i := 0; i < 2; i++ {
		if err := <-errorChannel; err != nil {
			return err, nil
		}
	}
	log.L().Debug("All Goroutines exited normally")
	return nil, f.result
}

func (f *baseFinder) walkDirectory() chan walkDirectoryYield {
	log.L().Debug("Starting walking through directory")
	channel := make(chan walkDirectoryYield)

	go func() {
		defer close(channel)
		_ = filepath.Walk(f.targetDirectory, func(path string, info os.FileInfo, err error) error {
			return f.processWalkEntry(path, info, err, channel)
		})
	}()

	return channel
}

func (f *baseFinder) processWalkEntry(path string, info os.FileInfo, err error, channel chan<- walkDirectoryYield) error {
	if err != nil {
		return f.handleWalkError(path, err, channel)
	}
	if info.IsDir() {
		log.L().Debug("Discovered directory", zap.String("name", info.Name()))
		return nil
	}
	if !f.fileFilter(path, info) {
		log.L().Debug("Skipped file (filtered)", zap.String("name", info.Name()))
		return nil
	}
	log.L().Debug("Discovered file", zap.String("name", info.Name()))
	channel <- walkDirectoryYield{nil, &FileInfo{
		Name: info.Name(),
		Path: path,
		Size: info.Size(),
		Hash: "",
	}}
	return nil
}

func (f *baseFinder) handleWalkError(path string, err error, channel chan<- walkDirectoryYield) error {
	wrappedErr := errors.Join(err, fmt.Errorf("error accessing path %q", path))
	channel <- walkDirectoryYield{wrappedErr, nil}
	return wrappedErr
}

func (f *baseFinder) groupDuplicates(fileInfo FileInfo) {
	existing, exists := f.result[fileInfo.Hash]
	if exists {
		log.L().Debug("Found duplicate", zap.String("hash", fileInfo.Hash))
	}
	f.result[fileInfo.Hash] = append(existing, fileInfo)
}

func (f *baseFinder) createHashFunction() pool.TaskFunction[taskInput, taskOutput] {
	return func(ctx context.Context, input taskInput) taskOutput {
		if input.fileInfo == nil {
			return taskOutput{nil, nil}
		}

		log.L().Info("Calculating hash", zap.String("name", input.fileInfo.Name))

		hash, err := f.hashFile(ctx, input.fileInfo)
		if err != nil {
			return taskOutput{input.fileInfo, err}
		}

		input.fileInfo.Hash = fmt.Sprintf("%x", hash)
		log.L().Debug("Hash calculated",
			zap.String("name", input.fileInfo.Name),
			zap.String("hash", input.fileInfo.Hash))

		return taskOutput{input.fileInfo, nil}
	}
}

func (f *baseFinder) hashFile(ctx context.Context, fileInfo *FileInfo) ([]byte, error) {
	select {
	case <-ctx.Done():
		log.L().Debug("Task function received cancelled signal")
		return nil, errors.New("task cancelled")
	default:
	}

	file, err := os.Open(fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", fileInfo.Path, err)
	}
	log.L().Debug("Opened file", zap.String("name", fileInfo.Name))
	defer func() {
		_ = file.Close()
		log.L().Debug("Closed file", zap.String("name", fileInfo.Name))
	}()

	hash, err := f.hasher.Hash(file)
	if err != nil {
		return nil, fmt.Errorf("failed to hash %q: %w", fileInfo.Path, err)
	}
	return hash, nil
}
