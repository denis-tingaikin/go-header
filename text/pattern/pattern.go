package pattern

import (
	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/text"
)

//Pattern means specific field handler. For example handler for field {email}
type Pattern interface {
	Name() string
	Verify(text.Reader) messages.ErrorList
}

//NewPatternFunc creates new pattern from func
func NewPatternFunc(name string, verify func(text.Reader) messages.ErrorList) Pattern {
	return &patternFunc{name: name, verify: verify}
}

type patternFunc struct {
	name   string
	verify func(text.Reader) messages.ErrorList
}

func (r *patternFunc) Name() string {
	return r.name
}

func (r *patternFunc) Verify(reader text.Reader) messages.ErrorList {
	if r.verify == nil {
		messages.NewErrorList(messages.VerifyFuncNotProvided())
	}
	return r.verify(reader)
}
