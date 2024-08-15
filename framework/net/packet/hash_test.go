package packet

import (
	"google.golang.org/protobuf/reflect/protoregistry"
	"hash/crc32"
	"testing"
)

func TestRegisterMessage(t *testing.T) {

	// 检查哈希冲突

	check := func(filePath string) {
		fd, err := protoregistry.GlobalFiles.FindFileByPath(filePath)
		if err != nil {
			t.Errorf("find file failed: %v", err)
			return
		}

		set := make(map[uint32]struct{})
		mds := fd.Messages()
		for i := mds.Len() - 1; i >= 0; i-- {
			x := mds.Get(i)
			fullName := x.FullName()
			_, err = protoregistry.GlobalTypes.FindMessageByName(fullName)
			if err != nil {
				t.Errorf("find message failed: %v", err)
				return
			}
			msgID := crc32.ChecksumIEEE([]byte(fullName))
			if _, exist := set[msgID]; exist {
				t.Fail()
			}
			set[msgID] = struct{}{}
		}
	}

	check("climsg.proto")
	check("svrmsg.proto")
}
