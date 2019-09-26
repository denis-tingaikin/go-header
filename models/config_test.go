package models

import (
	"regexp"
	"testing"

	"github.com/go-header/messages"
)

func TestConfig1(t *testing.T) {
	config := Configuration{}
	errs := config.Validate()

	if errs.Empty() {
		t.Fail()
	}

	epxected := messages.NoRules()

	if epxected.Error() != errs.String() {
		t.Fail()
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

	epxected := messages.IncorrectGoroutineCount(-1).Error() + "\n" + messages.NoRules().Error()

	if epxected != actual.String() {
		t.Fail()
	}
}

func TestConfig3(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{}},
	}
	actual := config.Validate()

	if actual.Empty() {
		t.Fail()
	}

	if actual.String() != messages.TemplateNotProvided().Error() {
		println(actual.String())
		t.Fail()
	}
}

func TestConfig4(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template: "Header...",
		}},
	}
	actual := config.Validate()

	if !actual.Empty() {
		t.Fail()
	}
}
func TestConfig5(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template:      "Header...",
			PathMatcher:   "[*]",
			AuthorMatcher: "[*Author1*]",
		}},
	}
	actual := config.Validate()

	if !actual.Empty() {
		t.Fail()
	}
}
func TestConfig6(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template:      "Header...",
			PathMatcher:   "*",
			AuthorMatcher: "*Author1*",
		}},
	}
	actual := config.Validate()
	if actual.Empty() {
		t.Fail()
	}
	_, err1 := regexp.Compile(config.Rules[0].PathMatcher)
	_, err2 := regexp.Compile(config.Rules[0].AuthorMatcher)
	expected := messages.CantProcessField(config.Rules[0].PathMatcher, err1).Error() + "\n" + messages.CantProcessField(config.Rules[0].AuthorMatcher, err2).Error()
	if actual.String() != expected {
		println(actual.String())
		println(expected)
		t.Fail()
	}
}
func TestConfig7(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template:      "Header...",
			PathMatcher:   "[*]",
			AuthorMatcher: "*Author1*",
		}},
	}
	actual := config.Validate()
	if actual.Empty() {
		t.Fail()
	}
	_, err := regexp.Compile(config.Rules[0].AuthorMatcher)
	expected := messages.CantProcessField(config.Rules[0].AuthorMatcher, err).Error()
	if actual.String() != expected {
		println(actual.String())
		println(expected)
		t.Fail()
	}
}
func TestConfig8(t *testing.T) {
	config := Configuration{
		Rules: []Rule{{
			Template:      "Header...",
			AuthorMatcher: ".*@company",
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
