package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	err := ReadIniFile("../config.ini")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%q\n", Addr)
	fmt.Printf("%v\n", NumCpu)
	fmt.Printf("%v\n", AcceptTimeout)
	fmt.Printf("%v\n", ReadTimeout)
	fmt.Printf("%v\n", WriteTimeout)
	fmt.Printf("%v\n", C2CRedisAddr)
	fmt.Printf("%v\n", C2CRedisKeyExpire)
	fmt.Printf("%v\n", GroupRedisAddr)
	fmt.Printf("%v\n", GroupRedisKeyExpire)
	fmt.Printf("%v\n", UuidGroupRedisAddr)
	fmt.Printf("%v\n", UuidGroupRedisKeyExpire)
	fmt.Printf("%v\n", GroupUuidRedisAddr)
	fmt.Printf("%v\n", GroupUuidRedisKeyExpire)
	fmt.Printf("%q\n", LogFile)
	fmt.Printf("%q\n", EmailServerAddr)
	fmt.Printf("%q\n", EmailServerPort)
	fmt.Printf("%q\n", EmailAccount)
	fmt.Printf("%q\n", EmailPassword)
	fmt.Printf("%v\n", EmailDuration)
	fmt.Printf("%q\n", EmailToList)

	s := strings.Split(EmailToList, " ")
	for _, v := range s {
		if v != "" {
			fmt.Printf("%q\n", v)
		}
	}
}
