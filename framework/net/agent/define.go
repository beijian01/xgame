package agent

import (
	"encoding/binary"
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
	"time"
)

const (
	ResponseFuncName = "response"
)

var (
	heartbeatTime                  = time.Second * 60 // second
	writeBacklog                   = 64               // backlog size
	endian        binary.ByteOrder = binary.BigEndian // big endian
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

func SetEndian(e binary.ByteOrder) {
	if e != nil {
		endian = e
	}
}

type pendingMsg struct {
	common *pb.MsgCommon
	msg    proto.Message
}
