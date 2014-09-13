package c2cmsg

import (
	"testing"
)

func TestOfflineMsg(t *testing.T) {
	Init()

	if GetMsgNum("jay") != 0 {
		t.Fatal()
	}

	AddMsg("jay", "hello")
	AddMsg("jay", "world")
	if GetMsgNum("jay") != 2 {
		t.Fatal()
	}

	msgs := GetMsgs("jay", 2)
	if msgs[0] != "hello" || msgs[1] != "world" {
		t.Fatal()
	}

	DeleteMsgs("jay", 1, 2)
	DeleteMsgs("jay", 1, 2)
	if GetMsgNum("jay") != 0 {
		t.Fatal()
	}

	AddMsg("doudou", "i")
	AddMsg("doudou", "fuck")
	AddMsg("doudou", "you")
	DeleteMsgs("doudou", 3, GetMsgNum("doudou"))
	if GetMsgNum("doudou") != 0 {
		t.Fatal()
	}

	msgs = GetMsgs("fuck", 1)
	if len(msgs) != 0 {
		t.Fatal()
	}
}
