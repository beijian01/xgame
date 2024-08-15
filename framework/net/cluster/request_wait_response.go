package cherryCluster

import (
	"google.golang.org/protobuf/proto"
	"sync/atomic"
)

type ResponseWaitMgr struct {
	pbChan map[uint32]chan proto.Message
	mid    atomic.Uint32
}

func newResponseWaitMgr() *ResponseWaitMgr {
	return &ResponseWaitMgr{
		pbChan: make(map[uint32]chan proto.Message),
	}
}

func (p *ResponseWaitMgr) WaitResponse(mid uint32) proto.Message {
	p.pbChan[mid] = make(chan proto.Message)
	resp := <-p.pbChan[mid] // 阻塞等待
	return resp
}
func (p *ResponseWaitMgr) NextMid() uint32 {
	p.mid.Add(1)
	return p.mid.Load()
}
