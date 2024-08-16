package cherryCluster

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
)

const (
	Name = "cluster_component"
)

type Component struct {
	cfacade.Component
	cfacade.ICluster

	DoOnAfterInit []func()
}

func New() *Component {
	return &Component{}
}

func (c *Component) Name() string {
	return Name
}

func (c *Component) Init() {
	c.ICluster = c.loadCluster()
	c.ICluster.Init()
}

func (c *Component) OnAfterInit() {
	for _, fn := range c.DoOnAfterInit {
		fn()
	}
}
func (c *Component) OnStop() {
	c.ICluster.Stop()
}

func (c *Component) loadCluster() cfacade.ICluster {
	return NewCluster(c.App())
}
