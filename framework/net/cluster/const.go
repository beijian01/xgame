package cherryCluster

import (
	"fmt"
)

const (
	remoteSubjectFormat = "cherry.%s.remote.%s.%s" // nodeType.nodeId
)

// getRemoteSubject remote message nats chan
func getRemoteSubject(prefix, nodeType, nodeId string) string {
	return fmt.Sprintf(remoteSubjectFormat, prefix, nodeType, nodeId)
}
