package packet

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/gansidui/chatserver/utils/aes"
	. "github.com/gansidui/chatserver/utils/convert"
)

// ---------------------------数据包构造------------------------------------
// （1）字节序：大端模式
// （2）数据包组成：包长 + 包体填充长度 + 类型 + 包体
// 包长：4字节，uint32，整个数据包的长度
// 包体填充长度：2字节，uint16，包体末尾的填充字节数
// 类型：4字节，uint32
// 包体：字节数组，[]byte

// 对包体进行AES加密，要求加密原数据（包体）的长度为16的倍数。
// 先用protobuf封装消息得到包体，再在后面填充0x00字符，直到包体长度
// 达到16的倍数。再进行AES加密。
// --------------------------------------------------------------------------

// 数据包的类型
const (
	PK_ClientLogin = uint32(iota)
	PK_ServerAcceptLogin
	PK_ClientLogout
	PK_ClientPing
	PK_C2CTextChat
)

type Packet struct {
	Len    uint32
	PadLen uint16
	Type   uint32
	Data   []byte
}

// 得到序列化后的Packet
func (this *Packet) GetBytes() (buf []byte) {
	buf = append(buf, Uint32ToBytes(this.Len)...)
	buf = append(buf, Uint16ToBytes(this.PadLen)...)
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
	pbDataLen := len(pbData)

	pac := &Packet{}
	pac.Type = dataType
	pac.PadLen = uint16(16 - pbDataLen%16)

	if pac.PadLen != 0 {
		pac.Data = make([]byte, pbDataLen+int(pac.PadLen))
		copy(pac.Data[:pbDataLen], pbData[0:])
	} else {
		pac.Data = pbData
	}

	// 对Data进行AES加密
	pac.Data, err = aes.AesEncrypt(pac.Data)
	if err != nil {
		return nil, fmt.Errorf("Pack error: %v", err.Error())
	}

	pac.Len = uint32(10 + len(pac.Data))

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

	pbDataLen := len(decryptData) - int(pac.PadLen)

	err = proto.Unmarshal(decryptData[:pbDataLen], pb.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}
	return nil
}
