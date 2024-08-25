package packet

import (
	"fmt"
	log "github.com/beijian01/xgame/framework/logger"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"hash/crc32"
	"sync"
)

var (
	rwMutex sync.RWMutex
	id2type = make(map[uint32]protoreflect.MessageType)
	type2id = make(map[protoreflect.MessageType]uint32)
	id2name = make(map[uint32]protoreflect.FullName)
)

func RegisterMessage(msg proto.Message) (uint32, error) {
	if msg == nil {
		return 0, fmt.Errorf("msg is nil")
	}
	msgName := proto.MessageName(msg)
	msgType := msg.ProtoReflect().Type()
	id := crc32.ChecksumIEEE([]byte(msgName))

	rwMutex.Lock()
	defer rwMutex.Unlock()

	if oldMsgTyp, exist := id2type[id]; exist {
		if oldMsgTyp != msgType {
			panic(fmt.Sprintf("哈希冲突 %s %s", oldMsgTyp.Descriptor().FullName(), msgType.Descriptor().FullName()))
		}
		return id, nil
	}

	id2type[id] = msgType
	type2id[msgType] = id
	id2name[id] = msgName

	log.Debugf("RegisterMessage %s %d", msgName, id)

	return id, nil
}

func MessageType(id uint32) (protoreflect.MessageType, bool) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	if msgType, exist := id2type[id]; exist {
		return msgType, true
	}

	return nil, false
}

func MessageID(msgType protoreflect.MessageType) (uint32, bool) {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	if id, exist := type2id[msgType]; exist {
		return id, true
	}

	return 0, false
}

func MessageName(id uint32) (protoreflect.FullName, bool) {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	if name, exist := id2name[id]; exist {
		return name, true
	}
	return "", false
}

func OnUnmarshal(id uint32, data []byte) (proto.Message, error) {
	if msgType, bHave := MessageType(id); bHave {
		msg := msgType.New().Interface()
		err := proto.Unmarshal(data, msg)
		return msg, err
	}
	return nil, fmt.Errorf("message %v not registered", id)
}

func OnMarshal(msg proto.Message) (uint32, []byte, error) {
	msgType := msg.ProtoReflect().Type()

	if msgID, bHave := MessageID(msgType); bHave {
		data, err := proto.Marshal(msg.(proto.Message))
		return msgID, data, err
	}

	if msgID, e := RegisterMessage(msg); e == nil {
		data, err := proto.Marshal(msg)
		return msgID, data, err
	}

	return 0, nil, fmt.Errorf("message %v register failed", msgType)
}
