package messages

import (
	"errors"
	"fmt"
)

//ErrorMsg returns const string if err is nil otherwise returns formated error
func ErrorMsg(err error) string {
	if err == nil {
		return "<error is nil>"
	}
	return fmt.Sprintf("Error: %v", err.Error())
}

func CanNotLoadTemplateFromFile(reason error) error {
	return fmt.Errorf("can not load template file. %v", ErrorMsg(reason))
}

func NoRules() error {
	return errors.New("no rules defined")
}

func IncorrectGoroutineCount(actual int) error {
	return fmt.Errorf("incorrect goroutine count. Actual: %v. Expected: value should be more than zero", actual)
}

func CantProcessField(name string, err error) error {
	return fmt.Errorf("can not process field: \"%v\". %v", name, ErrorMsg(err))
}

func TemplateNotProvided() error {
	return errors.New("template not provided")
}

func UnknownPattern(patternName string) error {
	return fmt.Errorf("template: unknown pattern %v", patternName)
}

func VerifyFuncNotProvided() error {
	return errors.New("verify func not provided")
}

func WrongYear() error {
	return errors.New("expected year range of file creation year to current year")
}
func AnalysisError(position int, reason error) error {
	return fmt.Errorf("position: %v, %v", position, ErrorMsg(reason))
}

func Diff(actual, expected interface{}) error {
	return fmt.Errorf("expected: %v, actual: %v", expected, actual)
}

func Missed(what string) error {
	return fmt.Errorf("missed: %v", what)
}

func NotExpected(what string) error {
	return fmt.Errorf("not expected text: %v", what)
}

func NoHeader() error {
	return errors.New("file has not license header")
}
