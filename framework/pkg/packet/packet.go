package packet

import (
	"bytes"
	"encoding/binary"
	cerr "github.com/beijian01/xgame/framework/error"
	"github.com/beijian01/xgame/pb"
	"google.golang.org/protobuf/proto"
	"io"
)

// |---commonLen---|---msgLen---|---common---|---msg---|
// |---head1-------|---head2----|---body1----|---body2-|

const (
	head1Len    = 2
	head2Len    = 2
	headLength  = head1Len + head2Len
	maxBody1Len = 1<<(head1Len*8) - 1
	maxBody2Len = 1<<(head2Len*8) - 1
)

var endian = binary.LittleEndian

func ReadMessage(reader io.Reader) (*pb.MsgCommon, proto.Message, error) {

	header, err := io.ReadAll(io.LimitReader(reader, headLength))
	if err != nil {
		return nil, nil, err
	}

	// if the header has no reader, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, nil, cerr.PacketConnectClosed
	}

	commonLen, msgLen, err := parseHead(header)
	if err != nil {
		return nil, nil, err
	}

	rawCommon, err := io.ReadAll(io.LimitReader(reader, int64(commonLen)))
	if err != nil {
		return nil, nil, err
	}

	common := &pb.MsgCommon{}
	err = proto.Unmarshal(rawCommon, common)
	if err != nil {
		return nil, nil, err
	}

	rawMsg, err := io.ReadAll(io.LimitReader(reader, int64(msgLen)))
	if err != nil {
		return nil, nil, err
	}

	msg, err := OnUnmarshal(common.Route, rawMsg)
	if err != nil {
		return nil, nil, err
	}

	return common, msg, nil
}

func parseHead(header []byte) (uint16, uint16, error) {

	if len(header) != headLength {
		return 0, 0, cerr.PacketInvalidHeader
	}

	bytesReader := bytes.NewReader(header)

	var err error
	var commonLen, msgLen uint16

	err = binary.Read(bytesReader, endian, &commonLen)
	if err != nil {
		return 0, 0, err
	}
	if commonLen > maxBody1Len {
		return 0, 0, cerr.PacketSizeExceed
	}

	err = binary.Read(bytesReader, endian, &msgLen)
	if err != nil {
		return 0, 0, err
	}
	if msgLen > maxBody2Len {
		return 0, 0, cerr.PacketSizeExceed
	}

	return commonLen, msgLen, nil
}

func PackMessage(common *pb.MsgCommon, msg proto.Message) ([]byte, error) {
	pkg := bytes.NewBuffer([]byte{})
	var (
		err               error
		rawCommon, rawMsg []byte
	)
	common.Route, rawMsg, err = OnMarshal(msg)
	if err != nil {
		return nil, err
	}
	if rawCommon, err = proto.Marshal(common); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, endian, uint16(len(rawCommon))); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, endian, uint16(len(rawMsg))); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, endian, rawCommon); err != nil {
		return nil, err
	}
	if err = binary.Write(pkg, endian, rawMsg); err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}
