package gmodel

import (
	"strconv"
	"strings"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	s := RandomBytes(5, 4)
	if len(s) != 5 {
		t.Fatal()
	}

	s = RandomBytes(1, 1)
	if len(s) != 1 {
		t.Fatal()
	}

	s = RandomBytes(8, 9)
	if len(s) > 9 || len(s) < 8 {
		t.Fatal()
	}
}

func TestStringKey(t *testing.T) {
	var a uint64 = 1
	s := GetStringKey(a)
	if len(s) != 15 {
		t.Fatal()
	}
	if !strings.HasPrefix(s, strings.Repeat("0", 14)) {
		t.Fatal()
	}

	a = 123456789123456
	s = GetStringKey(a)
	if len(s) != 15 {
		t.Fatal()
	}
	if strings.HasPrefix(s, "0") {
		t.Fatal()
	}

	a = 1234567891234567
	s = GetStringKey(a)
	if len(s) != 16 {
		t.Fatal()
	}
	if s != strconv.FormatUint(a, 10) {
		t.Fatal()
	}
}
