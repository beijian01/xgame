package xworker

import (
	log "github.com/beijian01/xgame/framework/logger"
	"sync"
	"sync/atomic"
	"time"
)

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
	if w.closed.Load() {
		return
	}
	w.len.Add(1)
	w.fs <- f
}

func (w *Worker) AfterPost(duration time.Duration, f func()) {
	time.AfterFunc(duration, func() {
		w.Post(f)
	})
}

func (w *Worker) Start() {
	w.finiWg.Add(1)

	go func() {
		defer w.finiWg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Error("worker run panic", r)
				w.Start() // 挂了重启
			}
		}()

		for f := range w.fs {
			w.len.Add(-1)
			f()
		}

	}()
}

func (w *Worker) Len() int32 {
	return w.len.Load()
}

func (w *Worker) Stop() {
	if w.closed.CompareAndSwap(false, true) {
		close(w.fs)
		w.finiWg.Wait()
	}
}
