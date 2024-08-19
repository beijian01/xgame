package main

import (
	"fmt"
	"github.com/beijian01/xgame/framework/net/packet"
	"github.com/beijian01/xgame/pb"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	// 服务器的地址，格式为 "ip:port"
	serverAddr := "127.0.0.1:20202"
	// 连接到服务器
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("连接服务器失败:", err)
		os.Exit(1)
	}
	defer conn.Close() // 函数退出前关闭连接
	packet.RegisterMessage(&pb.SRespLogin{})
	go func() {
		for {
			common, msg, err := packet.ReadMessage(conn)
			if err != nil {
				logrus.Error(err)
				continue
			}
			logrus.Info(common, msg)
		}
	}()

	mid := uint32(0)
	for {
		mid++
		// 构建要发送的消息
		data, err := packet.PackMessage(&pb.MsgCommon{
			Mid: mid,
		}, &pb.CReqLogin{Account: fmt.Sprintf("布什*戈门 %d", rand.Int())})
		if err != nil {
			logrus.Error(err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			logrus.Error(err)
			return
		}

		time.Sleep(3 * time.Second)
	}

}
