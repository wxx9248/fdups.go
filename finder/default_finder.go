package finder

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"fdups/log"
	"fdups/pool"
	"go.uber.org/zap"
)

type taskInput struct {
	fileInfo *FileInfo
}

type taskOutput struct {
	fileInfo *FileInfo
	err      error
}

type defaultFinder struct {
	targetDirectory string
	workerPool      pool.WorkerPool[taskInput, taskOutput]
	result          map[string][]FileInfo
}

func NewDefaultFinder(targetDirectory string) Finder {
	return &defaultFinder{
		targetDirectory: targetDirectory,
		workerPool:      pool.NewDefaultWorkerPool[taskInput, taskOutput](runtime.NumCPU()),
		result:          make(map[string][]FileInfo),
	}
}

func (f *defaultFinder) Find() (error, map[string][]FileInfo) {
	f.workerPool.Start()
	log.L().Debug("Worker pool started")
	defer func() {
		f.workerPool.Stop()
		log.L().Debug("Worker pool stopped")
	}()

	errorChannel := make(chan error)
	log.L().Debug("Error handling channel created")
	defer func() {
		close(errorChannel)
		log.L().Debug("Error handling channel closed")
	}()

	go func() {
		for item := range f.walkDirectory() {
			if item.err != nil {
				log.L().Error("Error received, aborting", zap.Error(item.err))
				errorChannel <- item.err
				return
			}
			_ = f.workerPool.Submit(pool.Task[taskInput, taskOutput]{
				TaskFunction: calculateHash,
				Input:        taskInput{fileInfo: item.fileInfo},
			})
			log.L().Debug("Hashing task submitted", zap.String("name", item.fileInfo.Name))
		}
		f.workerPool.CloseSubmit()
		log.L().Debug("All task submitted; Goroutine exit")
		errorChannel <- nil
	}()

	go func() {
	outer:
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
				switch event {
				case pool.EventAllTaskDone:
					log.L().Debug("All task processed; Goroutine exit")
					break outer
				}
			}
		}

		errorChannel <- nil
	}()

	i := 0
	for err := range errorChannel {
		if err != nil {
			return err, nil
		}
		i++
		if i >= 2 {
			log.L().Debug("All Goroutines exited normally")
			break
		}
	}
	return nil, f.result
}

func calculateHash(context context.Context, input taskInput) taskOutput {
	if input.fileInfo == nil {
		return taskOutput{nil, nil}
	}

	log.L().Info("Calculating hash", zap.String("name", input.fileInfo.Name))

	file, err := os.Open(input.fileInfo.Path)
	if err != nil {
		return taskOutput{input.fileInfo, err}
	}
	log.L().Debug("Opened file", zap.String("name", input.fileInfo.Name))
	defer func() {
		_ = file.Close()
		log.L().Debug("Closed file", zap.String("name", input.fileInfo.Name))
	}()

	select {
	case <-context.Done():
		log.L().Debug("Task function received cancelled signal")
		return taskOutput{nil, errors.New("task cancelled")}
	default:
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			return taskOutput{input.fileInfo, err}
		}
		input.fileInfo.Hash = fmt.Sprintf("%x", h.Sum(nil))
		log.L().Debug("Hash calculated", zap.String("name", input.fileInfo.Name), zap.String("hash", input.fileInfo.Hash))

		return taskOutput{input.fileInfo, nil}
	}
}

type walkDirectoryYield struct {
	err      error
	fileInfo *FileInfo
}

func (f *defaultFinder) walkDirectory() chan walkDirectoryYield {
	log.L().Debug("Starting walking through directory")

	channel := make(chan walkDirectoryYield)

	go func() {
		defer close(channel)
		_ = filepath.Walk(f.targetDirectory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				err = errors.Join(err, fmt.Errorf("Error accessing Path %q\n", path))
				channel <- walkDirectoryYield{err, nil}
				// Notify the caller for early abortion
				return err
			}
			if info.IsDir() {
				log.L().Debug("Discovered directory", zap.String("name", info.Name()))
				return nil
			}
			log.L().Debug("Discovered file", zap.String("name", info.Name()))
			channel <- walkDirectoryYield{nil, &FileInfo{
				info.Name(),
				path,
				info.Size(),
				"",
			}}
			return nil
		})
	}()

	return channel
}

func (f *defaultFinder) groupDuplicates(fileInfo FileInfo) {
	value, ok := f.result[fileInfo.Hash]
	if ok {
		log.L().Debug("Found duplicate", zap.String("hash", fileInfo.Hash))
		f.result[fileInfo.Hash] = append(value, fileInfo)
	} else {
		f.result[fileInfo.Hash] = []FileInfo{fileInfo}
	}
}
