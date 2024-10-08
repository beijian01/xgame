package facade

import (
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/pkg/packet"
	"github.com/beijian01/xgame/pb"

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
	}

	MemberListener func(member IMember) // MemberListener 成员增、删监听函数
)

type (
	ICluster interface {
		Init()
		SendBytes(nodeId string, data []byte) error
		ListenMessage(cbk any)
		PublishMsg(nodeId string, msg proto.Message) error                                            // 异步 RPC，仅通知，不需要回复
		RequestWait(nodeId string, req proto.Message, timeout time.Duration) (proto.Message, error)   // 同步阻塞 RPC ,请求/回复
		RequestAsync(nodeId string, req proto.Message, cbk func(resp proto.Message, err error)) error // 异步回调 RPC，请求/回复
		Stop()
		RegisterResponse(resp proto.Message)
		SetDefaultHandler(handler ReqMsgHandler)
	}

	ReqMsgHandler func(ext *Sender, req proto.Message)
)

type Sender struct {
	*pb.MsgCommon

	App IApplication
}

func (s *Sender) Resp(msg proto.Message) {
	//s
	local := &pb.MsgCommon{
		Mid:      s.Mid,
		Route:    s.Route,
		Sid:      s.Sid,
		SourceId: s.TargetId,
		TargetId: s.SourceId,
		Uid:      s.Uid,
	}

	data, err := packet.PackMessage(local, msg)
	if err != nil {
		log.Error(err)
		return
	}
	err = s.App.Cluster().SendBytes(local.TargetId, data)
	if err != nil {
		log.Error(err)
		return
	}
}

func (s *Sender) GetCommon() *pb.MsgCommon {
	return s.MsgCommon
}
