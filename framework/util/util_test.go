package util

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestTry(t *testing.T) {
	Try(func() {
		panic("panic here")
	}, func(errString string) {
		t.Log(errString)
		assert.Equal(t, "panic here", errString)
	})

	Try(func() {
		t.Log("no panic")
	}, func(errString string) {
		t.Log(errString)
	})
}
