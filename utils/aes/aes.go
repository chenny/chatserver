package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AES是对称加密
type AES struct {
	block cipher.Block
}

// 该包默认的密匙
const defaultAesKey = "12345abcdef67890"

var defaultAES *AES

func NewAES() *AES {
	return &AES{}
}

// 设置密匙
func (this *AES) SetAesKey(aesKey string) error {
	cblock, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return fmt.Errorf("aes.NewCipher: %v", err.Error())
	}
	this.block = cblock
	return nil
}

// AES加密
func (this *AES) AesEncrypt(src []byte) ([]byte, error) {
	// 验证输入参数
	// 原始数据长度需要为aes.BlockSize的整倍数， aes.BlockSize == 16
	if len(src)%aes.BlockSize != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	encryptText := make([]byte, aes.BlockSize+len(src))
	iv := encryptText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(this.block, iv)
	mode.CryptBlocks(encryptText[aes.BlockSize:], src)

	return encryptText, nil
}

// AES解密
func (this *AES) AesDecrypt(src []byte) ([]byte, error) {
	// hex
	decryptText, err := hex.DecodeString(fmt.Sprintf("%x", string(src)))
	if err != nil {
		return nil, err
	}

	// 长度不能小于aes.BlockSize
	if len(decryptText) < aes.BlockSize {
		return nil, errors.New("crypto/cipher: ciphertext too short")
	}

	iv := decryptText[:aes.BlockSize]
	decryptText = decryptText[aes.BlockSize:]

	// 验证输入参数
	// 原始数据长度需要为aes.BlockSize的整倍数， aes.BlockSize == 16
	if len(decryptText)%aes.BlockSize != 0 {
		return nil, errors.New("crypto/cipher: ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(this.block, iv)
	mode.CryptBlocks(decryptText, decryptText)

	return decryptText, nil
}

func init() {
	defaultAES = NewAES()
	err := defaultAES.SetAesKey(defaultAesKey)
	if err != nil {
		panic("defaultAES.SetAesKey: " + err.Error())
	}
}

func AesEncrypt(src []byte) ([]byte, error) {
	return defaultAES.AesEncrypt(src)
}

func AesDecrypt(src []byte) ([]byte, error) {
	return defaultAES.AesDecrypt(src)
}
