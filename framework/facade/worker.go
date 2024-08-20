package facade

import (
	log "github.com/beijian01/xgame/framework/logger"
	"sync"
	"sync/atomic"
	"time"
)

type IWorker interface {
	Post(f func())
	AfterPost(d time.Duration, f func())

	Run()
	Fini()
	WorkerLen() int32
}

type Worker struct {
	closed atomic.Bool
	finiWg sync.WaitGroup
	fs     chan func()
	len    atomic.Int32
}

func NewWorker(maxWorkerLen int) *Worker {
	if maxWorkerLen == 0 {
		maxWorkerLen = 1e3
	}
	w := &Worker{
		fs: make(chan func(), maxWorkerLen),
	}
	return w
}

func (w *Worker) Post(f func()) {
	w.len.Add(1)
	w.fs <- f
}

func (w *Worker) AfterPost(duration time.Duration, f func()) {
	time.AfterFunc(duration, func() {
		w.Post(f)
	})
}

func (w *Worker) Run() {
	w.finiWg.Add(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("worker run panic", r)
			}
			if !w.closed.Load() {
				w.Run()
			}
			w.finiWg.Done()
		}()
		for f := range w.fs {
			w.len.Add(-1)
			f()
		}
	}()
}

func (w *Worker) WorkerLen() int32 {
	return w.len.Load()
}

func (w *Worker) Fini() {
	if w.closed.CompareAndSwap(false, true) {
		close(w.fs)
		w.finiWg.Wait()
	}
}
