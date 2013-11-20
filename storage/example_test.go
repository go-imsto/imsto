package storage

import (
	"testing"
)

var (
	hash = "5cc163b92ab9b482b4486999d354f91e"
	id   = "5hos8aw6atq7kcpvn1gweaf4u"
)

func TestEntryId(t *testing.T) {
	new_id, err := NewEntryIdFromHash(hash)

	if err != nil {
		t.Fatal(err)
	}

	if new_id.id != id {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", new_id, id)
	}
}

const (
	t_salt  = "abcd"
	t_value = "test"
)

func TestApiToken(t *testing.T) {
	var (
		ver   = apiVer(0)
		appid = AppId(0)
		vc    = valueCate(0)
	)
	token, err := newToken(ver, appid, []byte(t_salt))
	if err != nil {
		t.Fatal(err)
	}

	token.SetValue([]byte(t_value), vc)
	// t.Logf("api token bins: %x", token.Binary())
	str := token.String()
	t.Logf("api token strs: %s", str)
	t.Logf("api token hash: %x, stamp: %d, value: %s", token.hash, token.stamp, token.GetValue())

	token, err = newToken(ver, appid, []byte(t_salt))
	if err != nil {
		t.Fatal(err)
	}

	var ok bool
	ok, err = token.VerifyString(str)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("token ok: %v", ok)
	value := string(token.GetValue())
	t.Logf("token value: %s", value)

	if value != t_value {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", value, t_value)
	}
}
