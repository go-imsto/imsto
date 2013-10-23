package imsto

import (
	"testing"
)

var (
	hash = "5cc163b92ab9b482b4486999d354f91e"
	id   = "5hos8aw6atq7kcpvn1gweaf4u"
)

func TestEntryId(t *testing.T) {
	new_id, err := NewEntryId(hash)

	if err != nil {
		t.Fatal(err)
	}

	if string(new_id) != id {
		t.Fatalf("unexpected result from BaseConvert:\n+ %v\n- %v", new_id, id)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Logf("confDir: %v", GetConfDir())

	err := loadConfig("/opt/imsto/config")

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("confDir: %v", GetConfDir())
}

func TestMetaBrowse(t *testing.T) {
	t.Log("dsn: " + getConfig("imsto", "meta_dsn"))
	mw, err := NewMetaWrapper("")

	if err != nil {
		t.Fatal(err)
	}

	limit := 5
	offset := 0
	rows, err := mw.Browse(limit, offset)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(rows)
}
