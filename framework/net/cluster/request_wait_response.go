package cherryCluster

import (
	"errors"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
	"time"
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

func (p *ResponseWaitMgr) WaitResponse(mid uint32, timeout time.Duration) (proto.Message, error) {
	p.pbChan[mid] = make(chan proto.Message)

	select {
	case resp := <-p.pbChan[mid]:
		return resp, nil
	case <-time.After(timeout):
		return nil, errors.New("waitResponse timeout")

	}
}
func (p *ResponseWaitMgr) NextMid() uint32 {
	p.mid.Add(1)
	return p.mid.Load()
}
