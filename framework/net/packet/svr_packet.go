package packet

import (
	"bytes"
	"encoding/binary"
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
	"io"
	"net"

	cerr "github.com/beijian01/xgame/framework/error"
)

// [路由ID 4byte][消息体长度 2byte][扩展数据长度 2byte][消息体（介于0~65535 byte之间）][扩展数据]

const (
	SvrMsgHeadLength    = 4 + 2 + 2
	SvrMsgMaxDataLength = 1<<16 - 1
	SvrMsgMaxEXtLength  = 1<<16 - 1
)

var svrMsgEndian = binary.LittleEndian

type SvrMessage struct {
	Route  uint32 // crc32.CheckSum(proto.message.Name())-> Route ,字符串哈希
	MsgLen uint32
	ExtLen uint32
	RawMsg []byte
	RawExt []byte

	PBMsg proto.Message // RawMsg 反序列化后的结果
	PBExt *pb.SvrExtend // RawExt 反序列化后的结果
}

func ReadSvrMessage(conn net.Conn) (*SvrMessage, bool, error) {
	header, err := io.ReadAll(io.LimitReader(conn, SvrMsgHeadLength))
	if err != nil {
		return nil, true, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, true, cerr.PacketConnectClosed
	}

	msg, err := parseSvrMsgHeader(header)
	if err != nil {
		return nil, true, err
	}

	msg.RawMsg, err = io.ReadAll(io.LimitReader(conn, int64(msg.MsgLen)))
	if err != nil {
		return nil, true, err
	}
	err = proto.Unmarshal(msg.RawMsg, msg.PBMsg)
	if err != nil {
		return nil, true, err
	}

	msg.RawExt, err = io.ReadAll(io.LimitReader(conn, int64(msg.ExtLen)))
	if err != nil {
		return nil, true, err
	}
	err = proto.Unmarshal(msg.RawExt, msg.PBExt)
	if err != nil {
		return nil, true, err
	}

	return msg, false, nil
}

func parseSvrMsgHeader(header []byte) (*SvrMessage, error) {
	msg := &SvrMessage{}

	if len(header) != SvrMsgHeadLength {
		return msg, cerr.PacketInvalidHeader
	}

	bytesReader := bytes.NewReader(header)

	var err error

	err = binary.Read(bytesReader, svrMsgEndian, &msg.Route)
	if err != nil {
		return msg, err
	}

	err = binary.Read(bytesReader, svrMsgEndian, &msg.MsgLen)
	if err != nil {
		return msg, err
	}
	if msg.MsgLen > SvrMsgMaxDataLength {
		return msg, cerr.PacketSizeExceed
	}

	err = binary.Read(bytesReader, svrMsgEndian, &msg.ExtLen)
	if err != nil {
		return msg, err
	}
	if msg.ExtLen > SvrMsgMaxEXtLength {
		return msg, cerr.PacketSizeExceed
	}

	return msg, nil
}

func packSvrMsg(msg *SvrMessage) ([]byte, error) {
	pkg := bytes.NewBuffer([]byte{})
	var err error

	if err = binary.Write(pkg, svrMsgEndian, msg.Route); err != nil {
		return nil, err
	}

	if msg.RawMsg, err = proto.Marshal(msg.PBMsg); err != nil {
		return nil, err
	}

	if msg.RawExt, err = proto.Marshal(msg.PBExt); err != nil {
		return nil, err
	}

	if err = binary.Write(pkg, svrMsgEndian, uint16(len(msg.RawMsg))); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, svrMsgEndian, uint16(len(msg.RawExt))); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, svrMsgEndian, msg.RawMsg); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, svrMsgEndian, msg.RawExt); err != nil {
		return nil, err
	}

	return pkg.Bytes(), nil
}
