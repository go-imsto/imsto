package thumbs

import "fmt"

// CodeError ...
type CodeError struct {
	Code int
	Text string
	Path string
}

// Error ...
func (ie *CodeError) Error() string {
	return fmt.Sprintf("%d: %s", ie.Code, ie.Text)
}

// NewCodeError ...
func NewCodeError(code int, text string) *CodeError {
	return &CodeError{Code: code, Text: text}
}
