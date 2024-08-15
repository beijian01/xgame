package packet

import (
	"bytes"
	"encoding/binary"
	"google.golang.org/protobuf/proto"
	"io"
	"net"

	cerr "github.com/beijian01/xgame/framework/error"
)

// [消息唯一ID 4byte][路由ID 4byte][消息体长度 2byte][消息体（介于0~65535 byte之间）]

const (
	CLiMsgHeadLength    = 4 + 4 + 2
	CLiMsgMaxDataLength = 1<<16 - 1
)

var cliMsgEndian = binary.LittleEndian

type CliMessage struct {
	MID    uint32
	Route  uint32 // crc32.CheckSum(proto.message.Name())-> Route ,字符串哈希
	MsgLen uint16
	RawMsg []byte

	PBMsg proto.Message // RawMsg 反序列化后的结果
}

func ReadCliMessage(conn net.Conn) (*CliMessage, bool, error) {
	header, err := io.ReadAll(io.LimitReader(conn, CLiMsgHeadLength))
	if err != nil {
		return nil, true, err
	}

	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, true, cerr.PacketConnectClosed
	}

	msg, err := parseCliMsgHeader(header)
	if err != nil {
		return nil, true, err
	}

	msg.RawMsg, err = io.ReadAll(io.LimitReader(conn, int64(msg.MsgLen)))
	if err != nil {
		return nil, true, err
	}
	msg.PBMsg, err = OnUnmarshal(msg.Route, msg.RawMsg)
	if err != nil {
		return nil, true, err
	}

	return msg, false, nil
}

func parseCliMsgHeader(header []byte) (*CliMessage, error) {
	msg := &CliMessage{}

	if len(header) != CLiMsgHeadLength {
		return msg, cerr.PacketInvalidHeader
	}

	bytesReader := bytes.NewReader(header)

	err := binary.Read(bytesReader, cliMsgEndian, &msg.MID)
	if err != nil {
		return msg, err
	}

	err = binary.Read(bytesReader, cliMsgEndian, &msg.Route)
	if err != nil {
		return msg, err
	}

	err = binary.Read(bytesReader, cliMsgEndian, &msg.MsgLen)
	if err != nil {
		return msg, err
	}

	if msg.MsgLen > CLiMsgMaxDataLength {
		return msg, cerr.PacketSizeExceed
	}

	return msg, nil
}

func PackCliMsg(msg *CliMessage) ([]byte, error) {
	pkg := bytes.NewBuffer([]byte{})
	var err error

	if msg.Route, msg.RawMsg, err = OnMarshal(msg.PBMsg); err != nil {
		return nil, cerr.Wrap(err, "PackCliMsg OnMarshal")
	}
	msg.MsgLen = uint16(len(msg.RawMsg))
	if err = binary.Write(pkg, cliMsgEndian, msg.MID); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, cliMsgEndian, msg.Route); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, cliMsgEndian, msg.MsgLen); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, cliMsgEndian, msg.RawMsg); err != nil {
		return nil, err
	}

	return pkg.Bytes(), nil
}
