package cherryCluster

import (
	"fmt"
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type (
	MessageHandlerMgr struct {
		defaultHandler cfacade.ReqMsgHandler
		reqHandlers    map[uint32]cfacade.ReqMsgHandler // key=消息名的哈希值。请求消息对应的handler都是明确固定的，在启动时注册（map写操作），运行时只有读操作，所以不用加锁

		requester *requester
	}
)

func NewMessageHandlerMgr(app cfacade.IApplication) *MessageHandlerMgr {
	return &MessageHandlerMgr{
		reqHandlers: make(map[uint32]cfacade.ReqMsgHandler),
		requester:   newRpcHandlerMgr(app),
		defaultHandler: func(ext cfacade.ISender, req proto.Message) {
			logrus.Error("defaultHandler 消息未注册", req)
		},
	}
}

func (p *MessageHandlerMgr) ListenMsg(cbk any) {
	v := reflect.ValueOf(cbk)

	if v.Kind() != reflect.Func {
		panic("cbk is not func")
	}
	if v.Type().NumIn() != 2 {
		panic("cbk num in is not 2")
	}

	if !v.Type().In(0).Implements(reflect.TypeOf((*cfacade.ISender)(nil)).Elem()) {
		panic(fmt.Errorf("cbk param 0 is not ISender %v %v ", v.Type().In(0), reflect.TypeOf((*cfacade.ISender)(nil)).Elem()))
	}
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)

	id, err := packet.RegisterMessage(msg)
	if err != nil {
		logrus.Panic(err)
	}

	p.reqHandlers[id] = func(sender cfacade.ISender, req proto.Message) {
		v.Call([]reflect.Value{reflect.ValueOf(sender), reflect.ValueOf(req)})
	}
}

func (p *MessageHandlerMgr) RegisterResponse(resp proto.Message) {
	id, err := packet.RegisterMessage(resp)
	if err != nil {
		logrus.Panic(err)
	}

	p.reqHandlers[id] = func(sender cfacade.ISender, msg proto.Message) {
		cbk := p.requester.getCallback(sender.GetCommon().Mid)
		cbk(msg, nil)
	}
}
func (p *MessageHandlerMgr) setDefaultHandler(handler cfacade.ReqMsgHandler) {
	p.defaultHandler = handler
}
