package xgame

import (
	"fmt"
	"github.com/beijian01/xgame/framework/facade"
	log "github.com/beijian01/xgame/framework/logger"
	xcluster "github.com/beijian01/xgame/framework/pkg/cluster"
	xdiscovery "github.com/beijian01/xgame/framework/pkg/discovery"
	xworker "github.com/beijian01/xgame/framework/pkg/worker"
	"github.com/beijian01/xgame/framework/pkg/xagent"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/framework/util"

	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type (
	Application struct {
		facade.INode
		facade.IWorker
		isFrontend    bool
		running       int32               // is running
		dieChan       chan bool           // wait for end application
		onAfterInitFn []func()            // on after init execute functions
		onShutdownFn  []func()            // on shutdown execute functions
		components    []facade.IComponent // all components
		discovery     facade.IDiscovery   // discovery component
		cluster       facade.ICluster     // cluster component
		netParser     facade.INetParser   // pkg packet agent
		conf          *profile.ClusterCfg
	}
)

func (a *Application) Profile() *profile.ClusterCfg {
	return a.conf
}

func NewAppNode(conf *profile.ClusterCfg, nodeId string) (*Application, error) {
	nodeCfg, ok := conf.FindNode(nodeId)
	if !ok {
		return nil, fmt.Errorf("[nodeId = %s] not found", nodeId)
	}
	app := &Application{
		INode:      nodeCfg,
		isFrontend: nodeCfg.IsGate,
		dieChan:    make(chan bool),
		conf:       conf,
	}

	// 注册通用组件
	// worker：所有handler将放到worker的特定协程中顺序执行
	worker := xworker.New(1 << 10)
	app.Register(worker)
	app.IWorker = worker

	// 集群RPC
	cluster := xcluster.New(app)
	app.Register(cluster)
	app.SetCluster(cluster)

	// 服务发现
	discovery := xdiscovery.New()
	app.Register(discovery)
	app.SetDiscovery(discovery)

	// 如果是前端（网关）节点，则需要对客户端消息进行代理转发
	if app.IsFrontend() {
		// 消息解析
		netParser := xagent.NewNetParser(app)
		app.Register(netParser)
		app.SetNetParser(netParser)
		// 客户端代理
		agents := xagent.NewAgents()
		app.Register(agents)
	}

	// 其他特殊组件（仅特定服务需要的组件）由特定服务再自行注册

	return app, nil
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
			log.Errorf("[component = %T] name is nil", c)
			return
		}

		result := a.Find(c.Name())
		if result != nil {
			log.Errorf("[component name = %s] is duplicate.", c.Name())
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
		log.Error("Application has running.")
		return
	}

	defer log.Flush()

	// add connector component
	if a.netParser != nil {
		for _, connector := range a.netParser.Connectors() {
			a.Register(connector)
		}
	}

	log.Info("-------------------------------------------------")
	log.Infof("[nodeId      = %s] application is starting...", a.GetNodeId())
	log.Infof("[nodeType    = %s]", a.GetNodeType())
	log.Infof("[pid         = %d]", os.Getpid())

	// component list
	for _, c := range a.components {
		c.Set(a)
		log.Infof("[component = %s] is added.", c.Name())
	}
	log.Info("-------------------------------------------------")

	// execute Init()
	for _, c := range a.components {
		log.Infof("[component = %s] -> OnInit().", c.Name())
		c.Init()
	}
	log.Info("-------------------------------------------------")

	// execute OnAfterInit()
	for _, c := range a.components {
		log.Infof("[component = %s] -> OnAfterInit().", c.Name())
		c.OnAfterInit()
	}

	log.Info("-------------------------------------------------")

	a.IWorker.Run()
	// set application is running
	atomic.AddInt32(&a.running, 1)

	sg := make(chan os.Signal, 1)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	select {
	case <-a.dieChan:
		log.Info("invoke shutdown().")
	case s := <-sg:
		log.Infof("receive shutdown signal = %v.", s)
	}

	// stop status
	atomic.StoreInt32(&a.running, 0)

	log.Info("------- application will shutdown -------")

	if a.onShutdownFn != nil {
		for _, f := range a.onShutdownFn {
			util.Try(func() {
				f()
			}, func(errString string) {
				log.Warnf("[onShutdownFn] error = %s", errString)
			})
		}
	}

	//all components in reverse order
	for i := len(a.components) - 1; i >= 0; i-- {
		util.Try(func() {
			log.Infof("[component = %s] -> OnBeforeStop().", a.components[i].Name())
			a.components[i].OnBeforeStop()
		}, func(errString string) {
			log.Warnf("[component = %s] -> OnBeforeStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	for i := len(a.components) - 1; i >= 0; i-- {
		util.Try(func() {
			log.Infof("[component = %s] -> OnStop().", a.components[i].Name())
			a.components[i].OnStop()
		}, func(errString string) {
			log.Warnf("[component = %s] -> OnStop(). error = %s", a.components[i].Name(), errString)
		})
	}

	log.Info("------- application has been shutdown... -------")
}

func (a *Application) Shutdown() {
	a.IWorker.Fini()
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
