package storage

import (
	"testing"
)

func TestHstoreNew(t *testing.T) {
	text := `"ext"=>"", "size"=>"34508", "width"=>"758", "height"=>"140", "quality"=>"93", "no"=>NULL`

	var h = make(Hstore)
	err := h.Scan(text)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(h)

	// t.Log(h.String())

	t.Fail()
}

type testHstoreStruct struct {
	Name string
	Age  uint16
}

func TestStructToHstore(t *testing.T) {
	i := testHstoreStruct{"test name", uint16(23)}
	h := structToHstore(i)

	t.Log(h)

	t.Fail()
}
