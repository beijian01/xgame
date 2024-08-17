package cherryCluster

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type (
	MessageHandlerMgr struct {
		reqHandlers map[uint32]cfacade.ReqMsgHandler // key=消息名的哈希值。请求消息对应的handler都是明确固定的，在启动时注册（map写操作），运行时只有读操作，所以不用加锁
	}
)

func NewMessageHandlerMgr() *MessageHandlerMgr {
	return &MessageHandlerMgr{
		reqHandlers: make(map[uint32]cfacade.ReqMsgHandler),
	}
}

func (p *MessageHandlerMgr) ListenMsg(cbk any) {
	v := reflect.ValueOf(cbk)

	if v.Kind() != reflect.Func {
		logrus.Panic("cbk is not func")
	}
	if v.Type().NumIn() != 2 {
		logrus.Panic("cbk num in is not 2")
	}

	var common pb.MsgCommon
	if v.Type().In(0) != reflect.TypeOf(&common) {
		logrus.Panic("handler num in 0 is not MsgCommon")
	}
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)

	id, err := packet.RegisterMessage(msg)
	if err != nil {
		logrus.Panic(err)
	}

	p.reqHandlers[id] = func(ext *pb.MsgCommon, req proto.Message) {
		v.Call([]reflect.Value{reflect.ValueOf(ext), reflect.ValueOf(req)})
	}
}
