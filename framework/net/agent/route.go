package agent

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
			// 发生了哈希冲突，两条或两条以上不同的proto消息名映射到了同一个 uint32 上
			// 这是比较严重的问题，直接Fatal让程序中断退出
			// 实际上哈希冲突的概率小到可以忽略不计。如果真的发生了，就换个哈希函数或者修改发生冲突的消息名称。
			logrus.Fatalf("duplicate route.%v and %v have the same hash value",
				rm.route2Name, name)
		} else {
			logrus.Warnf("duplicate route %s", name)
		}
	}
	rm.route2nodeTyp[route] = nodeTyp
}
