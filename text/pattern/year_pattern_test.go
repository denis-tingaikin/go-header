package pattern

import (
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
	expected := messages.NewErrorList(messages.AnalysisError(0, messages.CatNotParseAsYear()))
	if !expected.Equals(errs) {
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
	expected := messages.NewErrorList(messages.AnalysisError(0, messages.WrongYear()))
	if !errs.Equals(expected) {
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
	expected := messages.NewErrorList(messages.AnalysisError(0, messages.WrongYear()))
	if !expected.Equals(errs) {
		println(errs.String())
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
	expected := messages.NewErrorList(messages.AnalysisError(5, messages.WrongYear()))
	if !expected.Equals(errs) {
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
	expected := messages.NewErrorList(messages.AnalysisError(5, messages.CatNotParseAsYear()))
	if !expected.Equals(errs) {
		t.FailNow()
	}
}
