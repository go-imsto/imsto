package storage

import (
	"testing"
)

func TestHstoreNew(t *testing.T) {
	text := `"ext"=>"", "size"=>"34508", "width"=>"758", "height"=>"140", "quality"=>"93", "no"=>NULL`

	h, err := newHstore(text)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(h)

	// t.Log(h.String())

	t.Fail()
}
