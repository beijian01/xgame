package cherryCluster

import (
	cherryFacade "github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type Sender struct {
	*pb.MsgCommon

	app cherryFacade.IApplication
}

func (s *Sender) Resp(msg proto.Message) {
	//s
	local := &pb.MsgCommon{
		Mid:      s.Mid,
		MsgType:  pb.MsgType_SvrMsgTypResponseWait,
		Route:    s.Route,
		Sid:      s.Sid,
		SourceId: s.TargetId,
		TargetId: s.SourceId,
		Uid:      s.Uid,
	}

	data, err := packet.PackMessage(local, msg)
	if err != nil {
		logrus.Error(err)
		return
	}
	err = s.app.Cluster().SendBytes(local.TargetId, data)
	if err != nil {
		logrus.Error(err)
		return
	}
}
