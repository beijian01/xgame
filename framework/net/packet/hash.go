package packet

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"hash/crc32"
	"sync"
)

var (
	msgMutex   sync.RWMutex
	msgID2Typ  = make(map[uint32]protoreflect.MessageType)
	msgType2ID = make(map[protoreflect.MessageType]uint32)
	msgID2Name = make(map[uint32]protoreflect.FullName)
)

func RegisterMessage(msg proto.Message) (uint32, error) {
	msgName := proto.MessageName(msg)
	msgType := msg.ProtoReflect().Type()
	return RegisterMessageNameType(msgName, msgType)
}

func RegisterMessageNameType(msgName protoreflect.FullName, msgType protoreflect.MessageType) (uint32, error) {
	id := crc32.ChecksumIEEE([]byte(msgName))

	if msgType == nil {
		return id, fmt.Errorf("register message, message name:%v, type:%v is nil", msgName, msgType)
	}

	msgMutex.Lock()
	defer msgMutex.Unlock()

	if oldMsgTyp, exist := msgID2Typ[id]; exist {
		if oldMsgTyp != msgType {
			// 哈希冲突.
			return id, fmt.Errorf("register message,hash error, message name:%v, type:%v, old message name:%v, type:%v", msgName, msgType, msgID2Name[id], oldMsgTyp)
		}
		return id, nil
	}

	msgID2Typ[id] = msgType
	msgType2ID[msgType] = id
	msgID2Name[id] = msgName

	logrus.WithFields(logrus.Fields{
		"msgid":   id,
		"msgtype": msgType,
		"msgName": msgName,
	}).Debug("RegisterMessage")

	return id, nil
}

func MessageType(id uint32) (protoreflect.MessageType, bool) {
	msgMutex.RLock()
	defer msgMutex.RUnlock()

	if msgType, bHave := msgID2Typ[id]; bHave {
		return msgType, true
	}

	return nil, false
}

func MessageID(msgType protoreflect.MessageType) (uint32, bool) {
	msgMutex.Lock()
	defer msgMutex.Unlock()

	if msgID, bHave := msgType2ID[msgType]; bHave {
		return msgID, true
	}

	return 0, false
}

func MessageName(id uint32) protoreflect.FullName {
	msgMutex.Lock()
	defer msgMutex.Unlock()
	if name, bHave := msgID2Name[id]; bHave {
		return name
	}
	return protoreflect.FullName(fmt.Sprintf("msgid_%d not found", id))
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

	return 0, nil, fmt.Errorf("OnMarshal, message %v auto register failed", msgType)
}
