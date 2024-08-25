package facade

import (
	"time"
)

type IWorker interface {
	Post(f func())
	AfterPost(d time.Duration, f func())
	Run()
	Fini()
	WorkerLen() int32
}
