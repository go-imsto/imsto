package db

import (
	"testing"
)

func TestQarrayNew(t *testing.T) {

	a := Qarray{"k1", "k32", "cd"}

	t.Log(a)

	q := Qarray(a)

	t.Log(q.Value())

}
