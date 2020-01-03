package messages

import (
	"fmt"
	"strings"

	"github.com/denis-tingajkin/go-header/utils"
)

//ErrorList provides API for collecting errors
type ErrorList interface {
	Append(...error)
	Empty() bool
	Errors() []error
	Equals(ErrorList) bool
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
		_, _ = sb.WriteString(utils.MakeFirstLetterUppercase(err.Error()))
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

func (l *errorList) Equals(another ErrorList) bool {
	return l.String() == another.String()
}
