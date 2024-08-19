package xagent

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"hash/crc32"
)

type routeMgr struct {
	route2nodeTyp map[uint32]string
	route2Name    map[uint32]protoreflect.FullName
}

func newRouteMgr() *routeMgr {
	return &routeMgr{
		route2nodeTyp: make(map[uint32]string),
		route2Name:    make(map[uint32]protoreflect.FullName),
	}
}

func (rm *routeMgr) addRoute(msg proto.Message, nodeTyp string) {
	name := msg.ProtoReflect().Descriptor().FullName()
	route := crc32.ChecksumIEEE([]byte(name))
	if _, exist := rm.route2nodeTyp[route]; exist {
		if rm.route2Name[route] != name {
			logrus.Fatalf("duplicate route.%v and %v have the same hash value",
				rm.route2Name, name)
		} else {
			logrus.Warnf("duplicate route %s", name)
		}
	}
	rm.route2nodeTyp[route] = nodeTyp
}
