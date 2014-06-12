package aes

import (
	"fmt"
	"testing"
)

func TestAES(t *testing.T) {
	aes := NewAES()

	err := aes.SetAesKey("1234567890abcdef")
	if err != nil {
		t.Error(err)
	}

	src := "lijie helloworld"
	encodeBytes, err := aes.AesEncrypt([]byte(src))
	if err != nil {
		t.Error(err)
	}

	decodeBytes, err := aes.AesDecrypt(encodeBytes)
	if err != nil {
		t.Error(err)
	}

	if string(decodeBytes) != src {
		t.Error(err)
	}

	fmt.Println(string(decodeBytes))
}

func TestDafualtAES(t *testing.T) {
	src := "helloworld lijie"
	encodeBytes, err := AesEncrypt([]byte(src))
	if err != nil {
		t.Error(err)
	}

	decodeBytes, err := AesDecrypt(encodeBytes)
	if err != nil {
		t.Error(err)
	}

	if string(decodeBytes) != src {
		t.Error(err)
	}

	fmt.Println(string(decodeBytes))
}
