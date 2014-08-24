package c2c

import (
	"strconv"
	"testing"
)

func TestId2Msg(t *testing.T) {
	for i := 0; i < 100; i++ {
		addMsg(strconv.Itoa(i), "hello "+strconv.Itoa(i))
	}

	if getMsg("12") != "hello 12" {
		t.Fatal()
	}

	deleteMsg("12")
	if getMsg("12") == "hello 12" {
		t.Fatal()
	}

	if getMsg("999") != "" {
		t.Fatal()
	}

	for i := 0; i < 100; i++ {
		deleteMsg(strconv.Itoa(i))
	}
	if getMsg("1") != "" || getMsg("99") != "" {
		t.Fatal()
	}
}
