package redislist

import (
	"strconv"
	"testing"
)

func TestList1(t *testing.T) {
	m := NewRedisListHelper("127.0.0.1:6380", 7*24*3600)
	m.Init()
	defer m.Clean()

	if m.GetLength("go") != 0 {
		t.Fatal()
	}

	m.RPush("go", "world")
	m.LPush("go", "hello")
	m.RPush("go", "i")
	m.RPush("go", "love")
	if m.GetLength("go") != 4 {
		t.Fatal()
	}

	ss := m.GetRangeValues("go", 0, 3)
	if ss[0] != "hello" || ss[1] != "world" || ss[2] != "i" || ss[3] != "love" {
		t.Fatal()
	}

	m.DelRangeValues("go", 1, 2, 4)
	if m.GetLength("go") != 2 {
		t.Fatal()
	}

	ss = m.GetRangeValues("go", 0, 1)
	if ss[0] != "hello" || ss[1] != "love" {
		t.Fatal()
	}

	m.DelRangeValues("go", 0, m.GetLength("go")-1, m.GetLength("go"))
	if m.GetLength("go") != 0 {
		t.Fatal()
	}
}

func TestList2(t *testing.T) {
	m := NewRedisListHelper("127.0.0.1:6380", 5)
	m.Init()
	defer m.Clean()

	if m.GetLength("haha") != 0 {
		t.Fatal()
	}

	for i := 0; i < 1000; i++ {
		m.RPush("haha", "hello "+strconv.Itoa(i))
	}

	if m.GetIndexValue("haha", 3) != "hello 3" {
		t.Fatal()
	}

	if m.GetIndexValue("haha", 999) != "hello 999" {
		t.Fatal()
	}
}

func TestList3(t *testing.T) {
	m := NewRedisListHelper("127.0.0.1:6380", -1)
	m.Init()
	defer m.Clean()

	if m.GetLength("souga") != 0 {
		t.Fatal()
	}

	for i := 0; i < 1000; i++ {
		m.RPush("souga", "hello "+strconv.Itoa(i))
	}

	if m.GetLength("souga") != 1000 {
		t.Fatal()
	}

	m.DelKey("souga")

	if m.GetLength("souga") != 0 {
		t.Fatal()
	}

}

func TestList4(t *testing.T) {
	m := NewRedisListHelper("127.0.0.1:6380", -1)
	m.Init()
	defer m.Clean()

	for i := 0; i < 5; i++ {
		m.RPush("a", strconv.Itoa(i))
	}

	m.RPush("a", "3")

	if m.GetLength("a") != 6 || !m.ExistValue("a", "3") {
		t.Fatal()
	}

	m.DelValue("a", "3")

	if m.GetLength("a") != 5 || !m.ExistValue("a", "3") {
		t.Fatal()
	}

	ss := m.GetRangeValues("a", 0, 4)
	if ss[3] != "3" {
		t.Fatal()
	}

	m.DelKey("a")

	if m.GetLength("a") != 0 {
		t.Fatal()
	}
}
