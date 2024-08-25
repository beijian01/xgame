package facade

import (
	"time"
)

// IWorker 独立的工作协程
type IWorker interface {
	Post(f func())                       // 添加任务
	AfterPost(d time.Duration, f func()) // 定时添加任务
	Start()                              // 启动工作协程，顺序执行任务
	Stop()                               // 等待剩余任务完成再关闭
	Len() int32                          // 待执行的任务数量
}
