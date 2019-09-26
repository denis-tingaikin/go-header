package messages

import (
	"fmt"
	"strings"
)

//ErrorList provides API for collecting errors
type ErrorList interface {
	Append(...error)
	Empty() bool
	Errors() []error
	fmt.Stringer
}

type errorList struct {
	errors []error
}

//NewErrorList returns new error list
func NewErrorList(errs ...error) ErrorList {
	return &errorList{errors: errs}
}

func (l *errorList) String() string {
	sb := strings.Builder{}
	for i, err := range l.errors {
		if err == nil {
			continue
		}
		_, _ = sb.WriteString(err.Error())
		if i+1 < len(l.errors) {
			_, _ = sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (l *errorList) Append(errs ...error) {
	l.errors = append(l.errors, errs...)
}
func (l *errorList) Errors() []error {
	return l.errors

}
func (l *errorList) Empty() bool {
	return len(l.errors) == 0
}
