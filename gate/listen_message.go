package gate

import (
	"github.com/beijian01/xgame/framework/facade"
	"github.com/beijian01/xgame/pb"
	"hash/crc32"
)

func onCReqLogin(sender *facade.Sender, req *pb.CReqLogin) {
	resp := pb.SRespLogin{
		Code: 0,
		Uid:  uint64(crc32.ChecksumIEEE([]byte(req.Account))),
	}
	sender.Resp(&resp)
}
