package analysis

import (
	"context"
	"testing"

	"github.com/denis-tingajkin/go-header/messages"
	"github.com/denis-tingajkin/go-header/text"

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
	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{}, messages.Diff(actual, expected)))
	if !expectedResult.Equals(errs) {
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

	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 4}, messages.Diff("bc", "abc")))
	if !expectedResult.Equals(errs) {
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

	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 3}, messages.Missed(" abc")))
	if !expectedResult.Equals(errs) {
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
	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 3}, messages.NotExpected(" abc")))
	if !expectedResult.Equals(errs) {
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

	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 3}, messages.Diff("_", " ")))
	if !expectedResult.Equals(errs) {
		t.FailNow()
	}
}

func TestPattern1(t *testing.T) {
	a := NewFromConfig(testConfig())
	expected := "a {year}b"
	actual := "a 123123b"
	ctx := WithTemplate(context.Background(), expected)
	errs := a.Analyse(ctx, actual)

	if errs.Empty() {
		t.FailNow()
	}

	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 2}, messages.WrongYear()))
	if !expectedResult.Equals(errs) {
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
	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 2}, messages.UnknownPattern("unknown pattern")))
	if !expectedResult.Equals(errs) {
		t.FailNow()
	}
}

func TestCustomPattern1(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:    "my pattern",
		Pattern: "my text... {year}",
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
		Pattern: "my text... {year}",
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern))
	expected := "a {my pattern}b"
	actual := "a my text.!. 2007b"
	ctx := WithTemplate(context.Background(), expected)
	actaulResult := a.Analyse(ctx, actual)
	expectedResult := messages.NewErrorList(messages.AnalysisError(text.Location{0, 10}, messages.Diff("!", ".")))
	if !expectedResult.Equals(actaulResult) {
		println(actaulResult.String())
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
	actaulResult := a.Analyse(ctx, actual)
	expectedResult := messages.NewErrorList(
		messages.DetectedInfiniteRecursiveEntry(myPattern1.Name, myPattern2.Name, myPattern1.Name),
		messages.AnalysisError(text.Location{0, 13}, messages.NotExpected("...")))
	if !expectedResult.Equals(actaulResult) {
		println(actaulResult.String())
		t.FailNow()
	}
}

func TestCustomPattern4(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:          "my pattern",
		Pattern:       "my text... {year}\n",
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

func TestCustomPattern5(t *testing.T) {
	myPattern := models.CustomPattern{
		Name:          "my pattern",
		Pattern:       "my text... {year}\n",
		AllowMultiple: true,
	}
	a := NewFromConfig(testConfigWithPatterns(myPattern))
	expected := "a {my pattern}b"
	actual := "a my text... 2007\nmy text... 2005-asd\nb"
	ctx := WithTemplate(context.Background(), expected)
	actaulResult := a.Analyse(ctx, actual)
	expectedResult := messages.NewErrorList(
		messages.Ambiguous(
			messages.NewErrorList(messages.AnalysisError(text.Location{1, 0}, messages.Diff("my text... 2005-asd\nb", "b"))),
			messages.NewErrorList(messages.AnalysisError(text.Location{1, 16}, messages.CatNotParseAsYear()), messages.AnalysisError(text.Location{1, 16}, messages.Diff("asd\nb", "\n")))))
	if !expectedResult.Equals(actaulResult) {
		println(expectedResult.String())
		println(actaulResult.String())
		t.FailNow()
	}
}
