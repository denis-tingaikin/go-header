package pattern

import (
	"strings"
	"testing"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/models"

	"github.com/denis-tingajkin/go-header/text"
)

func yearSampleConfig(year int) models.ReadOnlyConfiguration {
	config := models.Configuration{
		Year: year,
	}
	return models.AsReadonly(&config)
}

func TestYearRange1(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2019")
	errs := rule.Verify(reader)
	if !errs.Empty() {
		t.FailNow()
	}
}

func TestYearRange2(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2018-2019")
	errs := rule.Verify(reader)
	if !errs.Empty() {
		t.Fail()
	}
}

func TestYearRange3(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("a")
	errs := rule.Verify(reader)
	if errs.Empty() && len(errs.Errors()) != 1 {
		t.FailNow()
	}
	actual := errs.String()
	if !(strings.Contains(actual, "position: 0") && strings.Contains(actual, "invalid syntax")) {
		t.FailNow()
	}
}

func TestYearRange4(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2020")
	errs := rule.Verify(reader)
	if errs.Empty() && len(errs.Errors()) != 1 {
		t.FailNow()
	}
	actual := errs.String()
	expected := messages.AnalysisError(0, messages.WrongYear()).Error()
	if actual != expected {
		t.FailNow()
	}
}

func TestYearRange5(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2018asd")
	errs := rule.Verify(reader)
	if errs.Empty() && len(errs.Errors()) != 1 {
		t.FailNow()
	}
	actual := errs.String()
	expected := messages.AnalysisError(0, messages.WrongYear()).Error()
	if actual != expected {
		t.FailNow()
	}
}

func TestYearRange6(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2018-2020")
	errs := rule.Verify(reader)
	if errs.Empty() && len(errs.Errors()) != 1 {
		t.FailNow()
	}
	actual := errs.String()
	expected := messages.AnalysisError(5, messages.WrongYear()).Error()
	if actual != expected {
		t.FailNow()
	}
}
func TestYearRange7(t *testing.T) {
	rule := YearRange(yearSampleConfig(2019))
	reader := text.NewReader("2018-asd")
	errs := rule.Verify(reader)
	if errs.Empty() && len(errs.Errors()) != 1 {
		t.FailNow()
	}
	actual := errs.String()
	if !(strings.Contains(actual, "position: 5") && strings.Contains(actual, "invalid syntax")) {
		t.FailNow()
	}
}
