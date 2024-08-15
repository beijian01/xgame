package cherryCluster

import (
	"google.golang.org/protobuf/proto"
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

// todo 异步回调RPC handler
// todo 消息类型区分
// 1 publish  2 requestWait  3 requestAsync  4 responseWait  5 responseAsync
