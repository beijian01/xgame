package xcluster

import (
	"fmt"
	"github.com/beijian01/xgame/framework/facade"
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/beijian01/xgame/framework/pkg/packet"

	"google.golang.org/protobuf/proto"
	"reflect"
)

type (
	MessageHandlerMgr struct {
		defaultHandler facade.ReqMsgHandler
		reqHandlers    map[uint32]facade.ReqMsgHandler // key=消息名的哈希值。请求消息对应的handler都是明确固定的，在启动时注册（map写操作），运行时只有读操作，所以不用加锁

		requester *requester
	}
)

func NewMessageHandlerMgr(app facade.IApplication) *MessageHandlerMgr {
	return &MessageHandlerMgr{
		reqHandlers: make(map[uint32]facade.ReqMsgHandler),
		requester:   newRpcHandlerMgr(app),
		defaultHandler: func(ext *facade.Sender, req proto.Message) {
			log.Error("defaultHandler 消息未注册", req)
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

	if v.Type().In(0) != reflect.TypeOf((*facade.Sender)(nil)) {
		panic(fmt.Sprint("cbk in 0 is not Sender", v.Type().In(0), reflect.TypeOf((*facade.Sender)(nil))))
	}
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)

	id, err := packet.RegisterMessage(msg)
	if err != nil {
		log.Panic(err)
	}

	p.reqHandlers[id] = func(sender *facade.Sender, req proto.Message) {
		v.Call([]reflect.Value{reflect.ValueOf(sender), reflect.ValueOf(req)})
	}
}

func (p *MessageHandlerMgr) RegisterResponse(resp proto.Message) {
	id, err := packet.RegisterMessage(resp)
	if err != nil {
		log.Panic(err)
	}

	p.reqHandlers[id] = func(sender *facade.Sender, msg proto.Message) {
		cbk := p.requester.getCallback(sender.GetCommon().Mid)
		cbk(msg, nil)
	}
}
func (p *MessageHandlerMgr) setDefaultHandler(handler facade.ReqMsgHandler) {
	p.defaultHandler = handler
}
