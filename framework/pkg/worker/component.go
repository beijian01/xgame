package xworker

import "github.com/beijian01/xgame/framework/facade"

const Name = "worker_component"

type Component struct {
	facade.Component
	facade.IWorker
}

func New(maxQueueLen int) *Component {
	return &Component{
		IWorker: NewWorker(maxQueueLen),
	}
}

func (c *Component) Name() string {
	return Name
}

func (c *Component) Init() {
}

func (c *Component) OnAfterInit() {
	c.IWorker.Run()
}
func (c *Component) OnStop() {
	c.IWorker.Fini()
}
