package base

import (
	"testing"
)

// func test(a string, frombase int, tobase int) {
// 	b, _ := BaseConvert(a, frombase, tobase)
// 	fmt.Printf("%s: %s\n", a, b)
// }

// func main() {
// 	test("5cc163b92ab9b482b4486999d354f91e", 16, 36)
// 	test("5hos8aw6atq7kcpvn1gweaf4u", 36, 16)
// }

var (
	a = "5cc163b92ab9b482b4486999d354f91e"
	b = "5hos8aw6atq7kcpvn1gweaf4u"
)

func TestEncode(t *testing.T) {

	str, err := BaseConvert(a, 16, 36)

	if err != nil {
		t.Fatal(err)
	}

	if str != b {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", str, b)
	}
	// test("5cc163b92ab9b482b4486999d354f91e", 16, 36)
}

func TestDecode(t *testing.T) {
	str, err := BaseConvert(b, 36, 16)

	if err != nil {
		t.Fatal(err)
	}

	if str != a {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", str, a)
	}
	// test("5hos8aw6atq7kcpvn1gweaf4u", 36, 16)
}
