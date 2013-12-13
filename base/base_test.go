package base

import (
	"testing"
)

type testpair struct {
	n, s string
	f, t int
}

var pairs = []testpair{
	{"", "", 1, 2},
	{"12", "12", 10, 10},
	{"5cc163b92ab9b482b4486999d354f91e", "5hos8aw6atq7kcpvn1gweaf4u", 16, 36},
	{"5cc163b92ab9b482b4486999d354f91e", "2P1FKmvE5PjCN4PhpocjBs", 16, 62},
	{"d1fb1bc11d2e992b4be5f770f35e345aa75a1d11", "oj0fiwpthr2v1ecf3p27ktokmh1a51t", 16, 36},
	{"d1fb1bc11d2e992b4be5f770f35e345aa75a1d11", "tXzMTsS55hMM9DEUf4BXSlq509b", 16, 62},
	{"d1fb1bc11d2e992b4be5f770f35e345aa75a1d1104", "1ZHZLT1diZProPNKUef2LeorGEBTu", 16, 62},
}

func testEqual(t *testing.T, msg string, args ...interface{}) bool {
	if args[len(args)-2] != args[len(args)-1] {
		t.Errorf(msg, args...)
		return false
	}
	return true
}

func TestEncode(t *testing.T) {
	for _, p := range pairs {
		got, err := BaseConvert(p.n, p.f, p.t)
		if err != nil {
			t.Error(err)
		} else {
			testEqual(t, "baseconvert(%q, %d, %d) = %q, want %q", p.n, p.f, p.t, got, p.s)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, p := range pairs {
		got, err := BaseConvert(p.s, p.t, p.f)
		if err != nil {
			t.Error(err)
		} else {
			testEqual(t, "baseconvert(%q, %d, %d) = %q, want %q", p.s, p.t, p.f, got, p.n)
		}
	}
}
