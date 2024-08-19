package xgame

import (
	"github.com/beijian01/xgame/framework/facade"
	xcluster "github.com/beijian01/xgame/framework/net/cluster"
	xdiscovery "github.com/beijian01/xgame/framework/net/discovery"
	"github.com/beijian01/xgame/framework/net/xagent"
	"github.com/beijian01/xgame/framework/util"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type (
	Application struct {
		facade.INode
		isFrontend    bool
		running       int32               // is running
		dieChan       chan bool           // wait for end application
		onAfterInitFn []func()            // on after init execute functions
		onShutdownFn  []func()            // on shutdown execute functions
		components    []facade.IComponent // all components
		discovery     facade.IDiscovery   // discovery component
		cluster       facade.ICluster     // cluster component
		netParser     facade.INetParser   // net packet agent
	}
)

func NewAppNode(node facade.INode, isFrontend bool) *Application {

	app := &Application{
		INode:      node,
		isFrontend: isFrontend,
		running:    0,
		dieChan:    make(chan bool),
	}

	cluster := xcluster.New(app)
	app.Register(cluster)
	app.SetCluster(cluster)

	discovery := xdiscovery.New()
	app.Register(discovery)
	app.SetDiscovery(discovery)

	if app.IsFrontend() {
		netParser := xagent.NewNetParser(app)
		app.Register(netParser)
		app.SetNetParser(netParser)

		agents := xagent.NewAgents()
		app.Register(agents)
	}

	return app
}

func (a *Application) IsFrontend() bool {
	return a.isFrontend
}

func (a *Application) Running() bool {
	return a.running > 0
}

func (a *Application) DieChan() chan bool {
	return a.dieChan
}

func (a *Application) Register(components ...facade.IComponent) {
	if a.Running() {
		return
	}

	for _, c := range components {
		if c == nil || c.Name() == "" {
			logrus.Errorf("[component = %T] name is nil", c)
			return
		}

		result := a.Find(c.Name())
		if result != nil {
			logrus.Errorf("[component name = %s] is duplicate.", c.Name())
			return
		}

		a.components = append(a.components, c)
	}
}

func (a *Application) Find(name string) facade.IComponent {
	if name == "" {
		return nil
	}

	for _, component := range a.components {
		if component.Name() == name {
			return component
		}
	}
	return nil
}

// Remove component by name
func (a *Application) Remove(name string) facade.IComponent {
	if name == "" {
		return nil
	}

	var removeComponent facade.IComponent
	for i := 0; i < len(a.components); i++ {
		if a.components[i].Name() == name {
			removeComponent = a.components[i]
			a.components = append(a.components[:i], a.components[i+1:]...)
			i--
		}
	}
	return removeComponent
}

func (a *Application) All() []facade.IComponent {
	return a.components
}

func (a *Application) OnShutdown(fn ...func()) {
	a.onShutdownFn = append(a.onShutdownFn, fn...)
}

// Startup load components before startup
func (a *Application) Startup() {
	if a.Running() {
		logrus.Error("Application has running.")
		return
	}

	// add connector component
	if a.netParser != nil {
		for _, connector := range a.netParser.Connectors() {
			a.Register(connector)
		}
	}

	logrus.Info("-------------------------------------------------")
	logrus.Infof("[nodeId      = %s] application is starting...", a.GetNodeId())
	logrus.Infof("[nodeType    = %s]", a.GetNodeType())
	logrus.Infof("[pid         = %d]", os.Getpid())

	// component list
	for _, c := range a.components {
		c.Set(a)
		logrus.Infof("[component = %s] is added.", c.Name())
	}
	logrus.Info("-------------------------------------------------")

	// execute Init()
	for _, c := range a.components {
		logrus.Infof("[component = %s] -> OnInit().", c.Name())
		c.Init()
	}
	logrus.Info("-------------------------------------------------")

	// execute OnAfterInit()
	for _, c := range a.components {
		logrus.Infof("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}

	logrus.Info("-------------------------------------------------")

	// set application is running
	atomic.AddInt32(&a.running, 1)

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-a.dieChan:
		logrus.Info("invoke shutdown().")
	case s := <-sg:
		logrus.Infof("receive shutdown signal = %v.", s)
	}

	// stop status
	atomic.StoreInt32(&a.running, 0)

	logrus.Info("------- application will shutdown -------")

	if a.onShutdownFn != nil {
		for _, f := range a.onShutdownFn {
			util.Try(func() {
				f()
			}, func(errString string) {
				logrus.Warnf("[onShutdownFn] error = %s", errString)
			})
		}
	}

	//all components in reverse order
	for i := len(a.components) - 1; i >= 0; i-- {
		util.Try(func() {
			logrus.Infof("[component = %s] -> OnBeforeStop().", a.components[i].Name())
			a.components[i].OnBeforeStop()
		}, func(errString string) {
			logrus.Warnf("[component = %s] -> OnBeforeStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	for i := len(a.components) - 1; i >= 0; i-- {
		util.Try(func() {
			logrus.Infof("[component = %s] -> OnStop().", a.components[i].Name())
			a.components[i].OnStop()
		}, func(errString string) {
			logrus.Warnf("[component = %s] -> OnStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	logrus.Info("------- application has been shutdown... -------")
}

func (a *Application) Shutdown() {
	a.dieChan <- true
}

func (a *Application) Discovery() facade.IDiscovery {
	return a.discovery
}

func (a *Application) Cluster() facade.ICluster {
	return a.cluster
}

func (a *Application) SetDiscovery(discovery facade.IDiscovery) {
	a.discovery = discovery
}

func (a *Application) SetCluster(cluster facade.ICluster) {
	a.cluster = cluster
}

func (a *Application) SetNetParser(netParser facade.INetParser) {
	a.netParser = netParser
}
