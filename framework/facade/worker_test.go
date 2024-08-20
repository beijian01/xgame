package facade

import (
	"github.com/go-playground/assert/v2"
	"testing"
	"time"
)

func TestWorker_Run(t *testing.T) {
	w := NewWorker(100)
	w.Run()
	w.Post(func() {
		t.Log(1)
	})

	w.Post(func() {
		t.Log(2)
	})

	w.Post(func() {
		panic(3)
	})

	ch := make(chan int, 1)
	w.Post(func() {
		t.Log(4)
		ch <- 4
	})

	var v int = -1

	select {
	case v = <-ch:
	case <-time.After(time.Second * 5):
	}

	// worker要保证panic之后自动恢复，不影响后续的任务
	assert.Equal(t, v, 4)
	w.Fini()
}
