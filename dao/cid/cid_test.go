package cid

import (
	"testing"
)

func TestUuid(t *testing.T) {
	if UuidCheckOnline("jay") {
		t.Error("not exist")
	}
	UuidOnLine("jay")
	if !UuidCheckOnline("jay") {
		t.Error("exist")
	}
	UuidOffLine("jay")
	if UuidCheckOnline("jay") {
		t.Error("not exist")
	}
}
