package image

// CountWriter ...
type CountWriter struct {
	n int
}

// Write implements for io.Writer
func (cw *CountWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	cw.n += n
	return
}

// Len return count value
func (cw *CountWriter) Len() int {
	return cw.n
}
