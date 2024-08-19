package xcluster

import (
	"github.com/beijian01/xgame/framework/facade"
)

const (
	Name = "cluster_component"
)

type Component struct {
	facade.Component
	facade.ICluster
}

func New(app facade.IApplication) *Component {
	return &Component{
		ICluster: newCluster(app),
	}
}

func (c *Component) Name() string {
	return Name
}

func (c *Component) Init() {
	c.ICluster.Init()
}

func (c *Component) OnAfterInit() {
}
func (c *Component) OnStop() {
	c.ICluster.Stop()
}
