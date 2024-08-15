package cherryDiscovery

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	clog "github.com/beijian01/xgame/framework/logger"
	cprofile "github.com/beijian01/xgame/framework/profile"
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
	config := cprofile.GetConfig("cluster").GetConfig("discovery")
	if config.LastError() != nil {
		logrus.Error("`cluster` property not found in profile file.")
		return
	}

	mode := config.GetString("mode")
	if mode == "" {
		logrus.Error("`discovery->mode` property not found in profile file.")
		return
	}

	discovery, found := discoveryMap[mode]
	if discovery == nil || !found {
		logrus.Errorf("mode = %s property not found in discovery config.", mode)
		return
	}

	logrus.Infof("Select discovery [mode = %s].", mode)
	p.IDiscovery = discovery
	p.IDiscovery.Load(p.App())
}

func (p *Component) OnStop() {
	p.IDiscovery.Stop()
}
