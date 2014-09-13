package groupinfo

import (
	"testing"
)

func TestGroupInfo(t *testing.T) {
	Init()

	ss := GetAllGroup("uuid")
	if len(ss) != 0 {
		t.Fatal()
	}

	ss = GetAllUuid("groupid")
	if len(ss) != 0 {
		t.Fatal()
	}

	_, groupid1 := BuildGroup("gansidui001", "lijie")
	_, groupid2 := BuildGroup("gansidui002", "lijie")
	ss = GetAllGroup("lijie")
	if len(ss) != 2 || ss[0] != groupid1 || ss[1] != groupid2 {
		t.Fatal()
	}

	group_name, group_owner := GetGroupNameAndOwner(groupid1)
	if group_name != "gansidui001" || group_owner != "lijie" {
		t.Fatal()
	}

	s2 := GetAllUuid(ss[0])
	if len(s2) != 2 || s2[0] != "gansidui001" || s2[1] != "lijie" {
		t.Fatal()
	}
	s2 = GetAllUuid(ss[1])
	if len(s2) != 2 || s2[0] != "gansidui002" || s2[1] != "lijie" {
		t.Fatal()
	}

	JoinGroup("doudou001", ss[0])
	JoinGroup("doudou001", ss[0])
	JoinGroup("doudou002", ss[0])
	JoinGroup("doudou003", ss[0])

	s2 = GetAllUuid(ss[0])
	if len(s2) != 5 || s2[0] != "gansidui001" || s2[1] != "lijie" ||
		s2[2] != "doudou001" || s2[3] != "doudou002" || s2[4] != "doudou003" {
		t.Fatal()
	}

	s2 = GetAllGroup("doudou002")
	if len(s2) != 1 || s2[0] != ss[0] {
		t.Fatal()
	}

	JoinGroup("doudou004", ss[0])

	ExitGroup("doudou002", ss[0])
	s2 = GetAllUuid(ss[0])
	if len(s2) != 5 || s2[0] != "gansidui001" || s2[1] != "lijie" ||
		s2[2] != "doudou001" || s2[3] != "doudou003" || s2[4] != "doudou004" {
		t.Fatal()
	}

	if !ExistUuidFromGroup(ss[0], "doudou003") {
		t.Fatal()
	}

}
