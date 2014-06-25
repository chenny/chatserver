package packet

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/chatserver/utils/aes"
	. "github.com/gansidui/chatserver/utils/convert"
)

// ---------------------------数据包构造------------------------------------
// （1）字节序：大端模式

// （2）数据包组成：包长 + 类型 + 包体

//  包长：4字节，uint32，整个数据包的长度

//  类型：4字节，uint32

//  包体：字节数组，[]byte

//  包长和类型用明文传输，包体由结构体采用protobuf序列化后再进行AES加密得到。
// --------------------------------------------------------------------------

// 数据包的类型
const (
	PK_ClientLogin       = uint32(1)
	PK_ServerAcceptLogin = uint32(2)
	PK_ClientLogout      = uint32(3)
	PK_ClientPing        = uint32(4)
	PK_C2CTextChat       = uint32(5)
)

type Packet struct {
	Len  uint32
	Type uint32
	Data []byte
}

// 得到序列化后的Packet
func (this *Packet) GetBytes() (buf []byte) {
	buf = append(buf, Uint32ToBytes(this.Len)...)
	buf = append(buf, Uint32ToBytes(this.Type)...)
	buf = append(buf, this.Data...)
	return buf
}

// 将数据包类型和pb数据结构一起打包成Packet，并加密Packet.Data
func Pack(dataType uint32, pb interface{}) (*Packet, error) {
	pbData, err := proto.Marshal(pb.(proto.Message))
	if err != nil {
		return nil, fmt.Errorf("Pack error: %v", err.Error())
	}

	pac := &Packet{}
	pac.Type = dataType

	// 对Data进行AES加密
	pac.Data, err = aes.AesEncrypt(pbData)
	if err != nil {
		return nil, fmt.Errorf("Pack error: %v", err.Error())
	}

	pac.Len = uint32(8 + len(pac.Data))

	return pac, nil
}

// 将Packet解包成非加密的pb数据结构
func Unpack(pac *Packet, pb interface{}) error {
	if pac == nil {
		return fmt.Errorf("Unpack error: pac == nil")
	}

	decryptData, err := aes.AesDecrypt(pac.Data)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	err = proto.Unmarshal(decryptData, pb.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}
	return nil
}
