package types

import (
	"testing"
)

func TestHstoreNew(t *testing.T) {
	text := `"name"=>"test", "ext"=>"", "size"=>"34508", "nov"=>NULL`

	var h = make(Hstore)
	err := h.Scan(text)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(h)

	valid_s := "34508"

	if h["size"] != valid_s {
		t.Fatalf("unexpected result:\n+ %v\n- %v", h["size"], valid_s)
	}
}

type testPerson struct {
	Name string
	Age  uint16
}

func TestStructToHstore(t *testing.T) {
	i := testPerson{"test name", uint16(23)}
	h := StructToHstore(i)

	t.Log(h)

	if h["name"] != i.Name {
		t.Fatalf("unexpected result:\n+ %v\n- %v", h["name"], i.Name)
	}
	if h["age"] != i.Age {
		t.Fatalf("unexpected result:\n+ %v\n- %v", h["age"], i.Age)
	}
}
