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

// func TestMetaBrowse(t *testing.T) {
// 	t.Log("dsn: " + config.GetValue("imsto", "meta_dsn"))
// 	mw, err := NewMetaWrapper("")

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	limit := 5
// 	offset := 0
// 	rows, err := mw.Browse(limit, offset)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Log(rows)
// }
