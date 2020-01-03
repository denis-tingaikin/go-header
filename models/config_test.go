package models

import (
	"testing"

	"github.com/denis-tingajkin/go-header/messages"
)

func TestConfig1(t *testing.T) {
	config := Configuration{}
	actual := config.Validate()

	if actual.Empty() {
		t.FailNow()
	}

	epxected := messages.NewErrorList(messages.NoRules())

	if !actual.Equals(epxected) {
		t.FailNow()
	}
}

func TestConfig2(t *testing.T) {
	config := Configuration{
		GoroutineCount: -1,
	}
	actual := config.Validate()

	if actual.Empty() {
		t.Fail()
	}

	epxected := messages.NewErrorList(messages.IncorrectGoroutineCount(-1), messages.NoRules())

	if !epxected.Equals(actual) {
		t.Fail()
	}
}

func TestConfig3(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{}},
		Scope: Scope{
			Policy: NonePolicy,
		},
	}
	actual := config.Validate()

	if actual.Empty() {
		t.Fail()
	}

	exepcted := messages.NewErrorList(messages.TemplateNotProvided())
	if !exepcted.Equals(actual) {
		println(actual.String())
		t.Fail()
	}
}

func TestConfig4(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
		}},
		Scope: Scope{
			Policy: NonePolicy,
		},
	}
	actual := config.Validate()

	if !actual.Empty() {
		println(actual.String())
		t.Fail()
	}
}
func TestConfig5(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
			Paths:    []string{"[*]"},
			Authors:  []string{"[*Author1*]"},
		}},
		Scope: Scope{
			Policy: NonePolicy,
		},
	}
	actual := config.Validate()

	if !actual.Empty() {
		t.Fail()
	}
}
func TestConfig6(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
			Paths:    []string{"*"},
			Authors:  []string{"*Author1*"},
		}},
		Scope: Scope{
			Policy: NonePolicy,
		},
	}
	actual := config.Validate()
	if actual.Empty() {
		t.Fail()
	}
	_, err1 := compileRegularExpressions(config.Rules[0].Authors)
	_, err2 := compileRegularExpressions(config.Rules[0].Paths)
	expected := messages.NewErrorList(err1, err2)
	if !actual.Equals(expected) {
		println(actual.String())
		println(expected.String())
		t.Fail()
	}
}
func TestConfig7(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
			Paths:    []string{"[*]"},
			Authors:  []string{"*Author1*"},
		}},
		Scope: Scope{
			Policy: NonePolicy,
		},
	}
	actual := config.Validate()
	if actual.Empty() {
		t.Fail()
	}
	_, err := compileRegularExpressions(config.Rules[0].Authors)
	expected := messages.NewErrorList(err)
	if !actual.Equals(expected) {
		println(actual.String())
		println(expected)
		t.Fail()
	}
}
func TestConfig8(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
			Authors:  []string{".*@company"},
		}},
	}
	s := Source{
		Author: "Author",
		Path:   "folder1",
	}
	config.Validate()
	actual := config.FindRule(&s)
	if actual != nil {
		t.FailNow()
	}
	s = Source{
		Author: "Author@company",
		Path:   "folder2",
	}
	actual = config.FindRule(&s)
	if actual == nil {
		t.FailNow()
	}
	if actual.Template != config.Rules[0].Template {
		t.FailNow()
	}
}
