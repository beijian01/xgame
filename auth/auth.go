package auth

import (
	xgame "github.com/beijian01/xgame/framework"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/profile"
	"github.com/beijian01/xgame/pb"
)

func Run(cfg *profile.ClusterCfg, nodeId string) error {
	node, err := xgame.NewAppNode(cfg, nodeId)
	if err != nil {
		return err
	}
	node.Cluster().ListenMessage(func(sender *facade.Sender, req *pb.ReqAuth) {
		resp := &pb.RespAuth{}
		sender.Resp(resp)
	})
	node.Startup()
	return nil
}
