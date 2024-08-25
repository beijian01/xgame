package xcluster

import (
	"fmt"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/pkg/packet"
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type (
	requester struct {
		sync.Mutex
		handlers map[uint32]requestCbk // key = mid
		mid      uint32

		app facade.IApplication
	}

	requestCbk func(response proto.Message, err error)
)

func newRpcHandlerMgr(app facade.IApplication) *requester {
	return &requester{
		handlers: make(map[uint32]requestCbk),
		app:      app,
	}
}

// 注册RPC的异步回调 （一次性，一个mid对应一个回调。回调执行完毕后注销）
func (p *requester) registerCallback(h requestCbk) uint32 {
	if h == nil {
		return 0
	}
	p.Lock()
	defer p.Unlock()
	p.mid++
	p.handlers[p.mid] = h
	return p.mid
}
func (p *requester) unregisterCallback(mid uint32) {
	p.Lock()
	defer p.Unlock()
	delete(p.handlers, mid)
}
func (p *requester) getCallback(mid uint32) requestCbk {
	p.Lock()
	defer p.Unlock()
	return func(response proto.Message, err error) {
		p.handlers[mid](response, err)
		p.unregisterCallback(mid)
	}
}

func (p *requester) requestAsync(nodeId string, req proto.Message, cbk func(resp proto.Message, err error)) error {
	mid := p.registerCallback(cbk)
	data, err := packet.PackMessage(&pb.MsgCommon{
		SourceId: p.app.GetNodeId(),
		TargetId: nodeId,
		Mid:      mid,
	}, req)
	if err != nil {
		return err
	}
	return p.app.Cluster().SendBytes(nodeId, data)
}

func (p *requester) requestWait(nodeId string, req proto.Message, timeout time.Duration) (proto.Message, error) {
	ch := make(chan proto.Message)
	mid := p.registerCallback(func(response proto.Message, err error) {
		ch <- response
	})
	data, err := packet.PackMessage(&pb.MsgCommon{
		SourceId: p.app.GetNodeId(),
		TargetId: nodeId,
		Mid:      mid,
	}, req)
	if err != nil {
		return nil, err
	}

	err = p.app.Cluster().SendBytes(nodeId, data)
	if err != nil {
		return nil, err
	}
	select {
	case <-time.After(timeout):
		return nil, fmt.Errorf("request wait timeout")
	case resp := <-ch:
		return resp, nil
	}

}
