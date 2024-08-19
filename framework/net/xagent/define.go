package xagent

import (
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
	"time"
)

var (
	heartbeatTime = time.Second * 60 // second
	writeBacklog  = 64               // backlog size
)

func SetHeartbeatTime(t time.Duration) {
	if t.Seconds() > 1 {
		heartbeatTime = t
	}
}

func SetWriteBacklog(backlog int) {
	if backlog > 0 {
		writeBacklog = backlog
	}
}

type pendingMsg struct {
	common *pb.MsgCommon
	msg    proto.Message
}
