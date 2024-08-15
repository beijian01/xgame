package cherryDiscovery

import (
	cfacade "github.com/beijian01/xgame/framework/facade"
	"github.com/sirupsen/logrus"
)

var (
	discoveryMap = make(map[string]cfacade.IDiscovery)
)

func init() {
	Register(&DiscoveryDefault{})
	//RegisterDiscovery(&DiscoveryETCD{})
}

func Register(discovery cfacade.IDiscovery) {
	if discovery == nil {
		logrus.Fatal("Discovery instance is nil")
		return
	}

	if discovery.Name() == "" {
		logrus.Fatalf("Discovery name is empty. %T", discovery)
		return
	}
	discoveryMap[discovery.Name()] = discovery
}
