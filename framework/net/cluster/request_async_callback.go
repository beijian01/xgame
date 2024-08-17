package cherryCluster

import (
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"reflect"
	"sync"
)

type (
	rpcCallbackMgr struct {
		sync.Mutex
		handlers map[uint32]rpcCallback // key = mid
		mid      uint32
	}

	rpcCallback func(response proto.Message, err error)
)

func newRpcHandlerMgr() *rpcCallbackMgr {
	return &rpcCallbackMgr{
		handlers: make(map[uint32]rpcCallback),
	}
}

// 注册RPC的异步回调 （一次性，一个mid对应一个回调。回调执行完毕后注销）
func (p *rpcCallbackMgr) registerCallback(h rpcCallback) {
	if h == nil {
		return
	}
	v := reflect.ValueOf(h)
	msg := reflect.New(v.Type().In(1)).Elem().Interface().(proto.Message)

	_, err := packet.RegisterMessage(msg)
	if err != nil {
		logrus.Panic(err)
	}

	p.Lock()
	defer p.Unlock()

	p.mid++
	p.handlers[p.mid] = h
}
func (p *rpcCallbackMgr) unregisterCallback(mid uint32) {
	p.Lock()
	defer p.Unlock()
	delete(p.handlers, mid)
}
func (p *rpcCallbackMgr) getCallback(mid uint32) rpcCallback {
	p.Lock()
	defer p.Unlock()
	return func(response proto.Message, err error) {
		p.handlers[mid](response, err)
		p.unregisterCallback(mid)
	}
}
