package cherryCluster

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"hash/crc32"
	"reflect"
)

type (
	MessageHandlerMgr struct {
		// 只有初始化时有写操作，之后的运行时都是读操作，所以不需要上锁

		cliHandlers map[uint32]cfacade.CliAgentHandler // key = route， 处理 client ->gate -> this server 的消息
		svrHandlers map[uint32]cfacade.SvrAgentHandler // 处理 server -> this server 的消息
	}
)

func NewMessageHandlerMgr() *MessageHandlerMgr {
	return &MessageHandlerMgr{
		cliHandlers: make(map[uint32]cfacade.CliAgentHandler),
		svrHandlers: make(map[uint32]cfacade.SvrAgentHandler),
	}
}

func (p *MessageHandlerMgr) ListenRequest(cbk any) {
	v := reflect.ValueOf(cbk)

	if v.Kind() != reflect.Func {
		logrus.Panic("ListenRequestSugar handler is not func")
	}
	// type check.
	if v.Type().NumIn() != 2 {
		logrus.Panic("ListenRequestSugar handler params num wrong")
	}
	var tempSender pb.SvrExtend
	if v.Type().In(0) != reflect.TypeOf(&tempSender) {
		logrus.Panic("ListenRequestSugar handler num in 0 is not Requester")
	}
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)
	name := proto.MessageName(msg)
	route := crc32.ChecksumIEEE([]byte(name))
	p.svrHandlers[route] = func(ext *pb.SvrExtend, msg proto.Message) {
		v.Call([]reflect.Value{reflect.ValueOf(ext), reflect.ValueOf(msg)})
	}
}

func (p *MessageHandlerMgr) ListenClientMsg(cbk cfacade.CliAgentHandler) {
	v := reflect.ValueOf(cbk)

	// type check.
	if v.Type().NumIn() != 2 {
		logrus.Panic("ListenRequestSugar handler params num wrong")
	}
	var tempSender pb.SvrExtend
	if v.Type().In(0) != reflect.TypeOf(&tempSender).Elem() {
		logrus.Panic("ListenRequestSugar handler num in 0 is not Requester")
	}
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)
	name := proto.MessageName(msg)
	route := crc32.ChecksumIEEE([]byte(name))
	p.cliHandlers[route] = cbk
}
