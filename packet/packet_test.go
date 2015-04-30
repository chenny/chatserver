package packet

import (
	"fmt"
	"github.com/gansidui/chatserver/pb"
	"github.com/gansidui/chatserver/utils/convert"
	proto "github.com/golang/protobuf/proto"
	"testing"
	"time"
)

func TestPacket(t *testing.T) {
	pbData := &pb.PbServerAcceptLogin{
		Login:     proto.Bool(true),
		TipsMsg:   proto.String("hello世界 哈哈，我爱你"),
		Timestamp: proto.Int64(time.Now().Unix()),
	}

	pac, err := Pack(PK_ServerAcceptLogin, pbData)
	if err != nil {
		t.Error(err)
	}

	ppd := &pb.PbServerAcceptLogin{}
	Unpack(pac, ppd)

	fmt.Println("Login:", ppd.GetLogin())
	fmt.Println("TipsMsg:", ppd.GetTipsMsg())
	fmt.Println("Timestamp:", convert.TimestampToTimeString(ppd.GetTimestamp()))
}
