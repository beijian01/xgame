package cherryDiscovery

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
)

const (
	Name = "discovery_component"
)

type Component struct {
	cfacade.Component
	cfacade.IDiscovery
}

func New() *Component {
	return &Component{}
}

func (*Component) Name() string {
	return Name
}

func (p *Component) Init() {
	p.IDiscovery = NewDiscoveryETCD()
	p.IDiscovery.Load(p.App())
}

func (p *Component) OnStop() {
	p.IDiscovery.Stop()
}
