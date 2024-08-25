package xworker

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestWorker_Run(t *testing.T) {
	w := NewWorker(100)
	w.Run()
	w.Post(func() {
		// 正常输出 1
		t.Log(1)
	})

	w.Post(func() {
		// 异常 除数为0
		var a, b = 1, 0
		t.Log(a / b)
	})

	w.Post(func() {
		// 异常
		panic("panic here , 3")
	})

	var v = -1
	w.Post(func() {
		// 正常输出 4
		t.Log(4)
		v = 4
	})

	w.Fini() //
	// worker要保证panic之后自动恢复，不影响后续的任务
	assert.Equal(t, v, 4) // 检查第四个任务是否正确执行了

}
