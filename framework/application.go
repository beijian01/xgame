package cherry

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/util"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type (
	Application struct {
		cfacade.INode
		isFrontend   bool
		running      int32                // is running
		dieChan      chan bool            // wait for end application
		onShutdownFn []func()             // on shutdown execute functions
		components   []cfacade.IComponent // all components
		discovery    cfacade.IDiscovery   // discovery component
		cluster      cfacade.ICluster     // cluster component
		netParser    cfacade.INetParser   // net packet parser
	}
)

func NewAppNode(node cfacade.INode, isFrontend bool) *Application {

	app := &Application{
		INode:      node,
		isFrontend: isFrontend,
		running:    0,
		dieChan:    make(chan bool),
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

func (a *Application) Register(components ...cfacade.IComponent) {
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

func (a *Application) Find(name string) cfacade.IComponent {
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
func (a *Application) Remove(name string) cfacade.IComponent {
	if name == "" {
		return nil
	}

	var removeComponent cfacade.IComponent
	for i := 0; i < len(a.components); i++ {
		if a.components[i].Name() == name {
			removeComponent = a.components[i]
			a.components = append(a.components[:i], a.components[i+1:]...)
			i--
		}
	}
	return removeComponent
}

func (a *Application) All() []cfacade.IComponent {
	return a.components
}

func (a *Application) OnShutdown(fn ...func()) {
	a.onShutdownFn = append(a.onShutdownFn, fn...)
}

// Startup load components before startup
func (a *Application) Startup() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Error(r)
		}
	}()

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

	// load net packet parser
	if a.isFrontend {
		if a.netParser == nil {
			logrus.Panic("net packet parser is nil.")
		}
		a.netParser.Load(a)
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

func (a *Application) Discovery() cfacade.IDiscovery {
	return a.discovery
}

func (a *Application) Cluster() cfacade.ICluster {
	return a.cluster
}

func (a *Application) SetDiscovery(discovery cfacade.IDiscovery) {
	if a.Running() || discovery == nil {
		return
	}

	a.discovery = discovery
}

func (a *Application) SetCluster(cluster cfacade.ICluster) {
	if a.Running() || cluster == nil {
		return
	}

	a.cluster = cluster
}

func (a *Application) SetNetParser(netParser cfacade.INetParser) {
	if a.Running() || netParser == nil {
		return
	}

	a.netParser = netParser
}
