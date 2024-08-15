package cherryFacade

type (
	// INode 节点信息
	INode interface {
		GetNodeId() string   // 节点id(全局唯一)
		GetNodeType() string // 节点类型
		GetAddress() string  // 对外网络监听地址(网关节点用)
	}

	IApplication interface {
		INode
		Running() bool                     // 是否运行中
		DieChan() chan bool                // die chan
		IsFrontend() bool                  // 是否为前端节点
		Register(components ...IComponent) // 注册组件
		Find(name string) IComponent       // 根据name获取组件对象
		Remove(name string) IComponent     // 根据name移除组件对象
		All() []IComponent                 // 获取所有组件列表
		OnShutdown(fn ...func())           // 关闭前执行的函数
		Startup()                          // 启动应用实例
		Shutdown()                         // 关闭应用实例
		Discovery() IDiscovery             // 发现服务
		Cluster() ICluster                 // 集群服务
	}
)
