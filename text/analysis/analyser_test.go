package analysis

import (
	"context"
	"strings"
	"testing"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/models"
)

func testConfig() models.ReadOnlyConfiguration {
	return models.AsReadonly(&models.Configuration{Year: 2007})
}

func testConfigWithPatterns(patterns ...models.CustomPattern) models.ReadOnlyConfiguration {
	return models.AsReadonly(&models.Configuration{Year: 2007, CustomPatterns: patterns})
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
		println(result)
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

func TestDiff6(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "abc abc"
	actual := "abc_abc"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	result := errs.String()
	expectedResult := messages.AnalysisError(3, messages.Diff("_", " ")).Error()
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

func TestCustomPattern1(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:    "my pattern",
		Pattern: "my text... {year range}",
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern))
	expected := "a {my pattern}b"
	actual := "a my text... 2007b"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if !errs.Empty() {
		t.FailNow()
	}
}

func TestCustomPattern2(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:    "my pattern",
		Pattern: "my text... {year range}",
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern))
	expected := "a {my pattern}b"
	actual := "a my text.!. 2007b"
	ctx := WithTemplate(context.Background(), expected)
	actaulResult := a.Analyse(ctx, actual).String()
	expectedResult := messages.AnalysisError(10, messages.Diff("!", ".")).Error()
	if actaulResult != expectedResult {
		t.FailNow()
	}
}

func TestCustomPattern3(t *testing.T) {
	myPattern1 := models.CustomPattern{
		Name:    "my pattern1",
		Pattern: "{my pattern2}",
	}
	myPattern2 := models.CustomPattern{
		Name:    "my pattern2",
		Pattern: "{my pattern1}",
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern1, myPattern2))
	expected := "{my pattern1}"
	actual := "..."
	ctx := WithTemplate(context.Background(), expected)
	actaulResult := a.Analyse(ctx, actual).String()
	expectedResult := messages.DetectedInfiniteRecursiveEntry(myPattern1.Name, myPattern2.Name, myPattern1.Name).Error()
	if !strings.Contains(actaulResult, expectedResult) {
		t.FailNow()
	}
}

func TestCustomPattern4(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:          "my pattern",
		Pattern:       "my text... {year range}\n",
		AllowMultiple: true,
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern))
	expected := "a {my pattern}b"
	actual := "a my text... 2007\nmy text... 2005-2007\nb"
	ctx := WithTemplate(context.Background(), expected)
	actaulResult := a.Analyse(ctx, actual)
	if !actaulResult.Empty() {
		println(actaulResult.String())
		t.FailNow()
	}
}
