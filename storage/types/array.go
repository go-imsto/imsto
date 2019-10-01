package types

import (
	"github.com/lib/pq"
)

type StringArray = pq.StringArray

type StringSlice pq.StringArray

// Index returns the index of the string s in StringSlice, or -1 if s is not found.
func (q StringSlice) Index(s string) int {
	for i, v := range q {
		if s == v {
			return i
		}
	}
	return -1
}

// Contains returns true if the string s is found
func (q StringSlice) Contains(s string) bool {
	return q.Index(s) > -1
}
