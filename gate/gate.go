package gate

import (
	xgame "github.com/beijian01/xgame/framework"
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/framework/net/profile"
	"github.com/beijian01/xgame/framework/net/xagent"
	"github.com/beijian01/xgame/pb"
)

//var conf = flag.String("conf", "profile.json", "config file")

func Run(cfg *profile.ClusterCfg, nodeId string) {
	nodeCfg, exist := cfg.FindNode(nodeId)
	if !exist {
		panic("node not found")
	}
	gate := xgame.NewAppNode(nodeCfg, true)
	gate.Startup()
}

func initForwardMessage(gate facade.IApplication) {
	agents := gate.Find(xagent.AgentsName).(*xagent.Agents)
	agents.RouteMessage((*pb.CReqLogin)(nil))
}
