package utils

import (
	"runtime/debug"
	"sync"

	"github.com/webcore-go/webcore/infra/logger"
)

const (
	backgroundWorkerCount = 15
	priorityWorkerCount   = 5
	backgroundQueueSize   = 1000
	priorityQueueSize     = 100
	backgroundOverflowMax = 100
)

type backgroundJob struct {
	name string
	fn   func()
}

var (
	backgroundQueue chan backgroundJob
	priorityQueue   chan backgroundJob
	overflowSem     chan struct{}
	poolOnce        sync.Once
)

func initBackgroundPool() {
	backgroundQueue = make(chan backgroundJob, backgroundQueueSize)
	priorityQueue = make(chan backgroundJob, priorityQueueSize)
	overflowSem = make(chan struct{}, backgroundOverflowMax)

	for i := 0; i < priorityWorkerCount; i++ {
		go priorityWorker()
	}
	for i := 0; i < backgroundWorkerCount; i++ {
		go backgroundWorker()
	}

	safeLogInfo("Background worker pool diinisialisasi",
		"workers", backgroundWorkerCount, "priority_workers", priorityWorkerCount,
		"queue_size", backgroundQueueSize, "priority_queue_size", priorityQueueSize,
		"overflow_max", backgroundOverflowMax)
}

func safeLogInfo(msg string, args ...any) {
	defer func() {
		_ = recover()
	}()
	logger.Info(msg, args...)
}

func safeLogError(msg string, args ...any) {
	defer func() {
		_ = recover()
	}()
	logger.Error(msg, args...)
}

func priorityWorker() {
	for job := range priorityQueue {
		runBackgroundJob(job)
	}
}
func backgroundWorker() {
	for {
		select {
		case job := <-priorityQueue:
			runBackgroundJob(job)
			continue
		default:
		}

		select {
		case job := <-priorityQueue:
			runBackgroundJob(job)
		case job := <-backgroundQueue:
			runBackgroundJob(job)
		}
	}
}

func runBackgroundJob(job backgroundJob) {
	defer RecoverBackground(job.name)
	safeLogInfo("Start Job: " + job.name + " di background")
	job.fn()
}
func RunBackground(name string, fn func()) {
	poolOnce.Do(initBackgroundPool)
	runBackgroundOnQueue(name, fn, backgroundQueue)
}
func RunBackgroundPriority(name string, fn func()) {
	poolOnce.Do(initBackgroundPool)
	runBackgroundOnQueue(name, fn, priorityQueue)
}

func runBackgroundOnQueue(name string, fn func(), queue chan backgroundJob) {
	select {
	case queue <- backgroundJob{name: name, fn: fn}:
		return
	default:
	}
	select {
	case overflowSem <- struct{}{}:
		go func() {
			defer func() { <-overflowSem }()
			defer RecoverBackground(name)
			safeLogInfo("Background overflow goroutine dipakai (antrian penuh)", "job", name)
			fn()
		}()
		return
	default:
	}
	safeLogInfo("Background pool & overflow penuh, menerapkan backpressure", "job", name)
	queue <- backgroundJob{name: name, fn: fn}
}

func RecoverBackground(name string) {
	if r := recover(); r != nil {
		safeLogError("Panic terjadi di background "+name, "err", r, "stack", string(debug.Stack()))
	}
}

func ResetBackgroundPoolForTest() {
	poolOnce = sync.Once{}
	backgroundQueue = nil
	priorityQueue = nil
	overflowSem = nil
}

func OverflowSemLenForTest() int {
	if overflowSem == nil {
		return 0
	}
	return len(overflowSem)
}
