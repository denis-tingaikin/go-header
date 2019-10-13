package models

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/denis-tingajkin/go-header/messages"
)

func TestRule1(t *testing.T) {
	filePath := path.Join(os.TempDir(), "testrule1.txt")
	expected := "header"
	err := ioutil.WriteFile(filePath, []byte(expected), os.ModePerm)
	if err != nil {
		t.FailNow()
	}
	defer func() {
		_ = os.Remove(filePath)
	}()

	rule := Rule{
		Template: expected,
	}
	if !rule.Compile().Empty() {
		t.FailNow()
	}
	if rule.Template != expected {
		t.FailNow()
	}
}

func TestRule2(t *testing.T) {
	rule := Rule{
		TemplatePath: "???",
	}
	actualErr := rule.loadTemplate()
	if actualErr == nil {
		t.FailNow()
	}
	actual := actualErr.Error()
	_, err := os.Open("???")
	expected := messages.CanNotLoadTemplateFromFile(err).Error()
	if actual != expected {
		println(actual)
		println(expected)
		t.FailNow()
	}
}
