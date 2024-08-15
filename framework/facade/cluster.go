package cherryFacade

import (
	"google.golang.org/protobuf/proto"
	"time"
)

type (
	// IDiscovery 发现服务接口
	IDiscovery interface {
		Load(app IApplication)
		Name() string                                                 // 发现服务名称
		Map() map[string]IMember                                      // 获取成员列表
		ListByType(nodeType string, filterNodeId ...string) []IMember // 根据节点类型获取列表
		Random(nodeType string) (IMember, bool)                       // 根据节点类型随机一个
		GetType(nodeId string) (nodeType string, err error)           // 根据节点id获取类型
		GetMember(nodeId string) (member IMember, found bool)         // 获取成员
		AddMember(member IMember)                                     // 添加成员
		RemoveMember(nodeId string)                                   // 移除成员
		OnAddMember(listener MemberListener)                          // 添加成员监听函数
		OnRemoveMember(listener MemberListener)                       // 移除成员监听函数
		Stop()
	}

	IMember interface {
		GetNodeId() string
		GetNodeType() string
		GetAddress() string
		GetSettings() map[string]string
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)

type (
	ICluster interface {
		Init()
		PublishMsg(nodeType string, msg proto.Message) error                                            // 异步 RPC，仅通知，不需要回复
		RequestWait(nodeType string, req proto.Message, timeout time.Duration) (proto.Message, error)   // 同步阻塞 RPC ,请求/回复
		RequestAsync(nodeType string, req proto.Message, cbk func(resp proto.Message, err error)) error // 异步 RPC，请求/回复
		Stop()                                                                                          // 停止
	}
)
