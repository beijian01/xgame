package cherryCluster

import (
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
)

type (
	MessageHandlerMgr struct {
		// 只有初始化时有写操作，之后的运行时都是读操作，所以不需要上锁

		cliHandlers map[uint32]CliAgentHandler // key = route
		svrHandlers map[uint32]SvrAgentHandler
	}

	CliAgentHandler func(session *pb.Session, req proto.Message)
	SvrAgentHandler func(session *pb.Session, req proto.Message)
)
