package analysis

import (
	"context"
	"testing"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/models"
)

func testConfig() models.ReadOnlyConfiguration {
	return models.AsReadonly(&models.Configuration{Year: 2007})
}

func TestDiff1(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc"
	actual := "bc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)
	if errs.Empty() || len(errs.Errors()) != 1 {
		t.FailNow()
	}
	result := errs.String()
	expectedResult := messages.AnalysisError(0, messages.Diff(actual, expected)).Error()
	if result != expectedResult {
		t.FailNow()
	}
}

func TestDiff2(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc"
	actual := "abc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if !errs.Empty() {
		t.FailNow()
	}
}

func TestDiff3(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc abc"
	actual := "abc bc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	result := errs.String()
	expectedResult := messages.AnalysisError(4, messages.Diff("bc", "abc")).Error()
	if result != expectedResult {
		t.FailNow()
	}
}

func TestDiff4(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc abc"
	actual := "abc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	result := errs.String()
	expectedResult := messages.AnalysisError(3, messages.Missed(" abc")).Error()
	if result != expectedResult {
		t.FailNow()
	}
}

func TestDiff5(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc"
	actual := "abc abc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	result := errs.String()
	expectedResult := messages.AnalysisError(3, messages.NotExpected(" abc")).Error()
	if result != expectedResult {
		t.FailNow()
	}
}

func TestPattern1(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "a {year range}b"
	actual := "a 123123b"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	result := errs.String()
	expectedResult := messages.AnalysisError(2, messages.WrongYear()).Error()
	if result != expectedResult {
		t.FailNow()
	}
}
func TestPattern2(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "a {unknown pattern}b"
	actual := "a b"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}
	result := errs.String()
	expectedResult := messages.AnalysisError(2, messages.UnknownPattern("unknown pattern")).Error()
	if result != expectedResult {
		t.FailNow()
	}
}
