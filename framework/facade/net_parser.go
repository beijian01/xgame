package facade

type (
	// INetParser 前端网络数据包解析器
	INetParser interface {
		AddConnector(connector IConnector)
		Connectors() []IConnector
	}
)
