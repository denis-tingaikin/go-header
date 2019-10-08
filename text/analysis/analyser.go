package analysis

import (
	"context"
	"strings"

	"github.com/denis-tingajkin/go-header/models"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/text"
	"github.com/denis-tingajkin/go-header/text/pattern"
)

type Analyser interface {
	Analyse(context.Context, string) messages.ErrorList
}

func New(patterns map[string]pattern.Pattern) Analyser {
	return &analyzer{patterns: patterns}
}

func NewFromConfig(config models.ReadOnlyConfiguration) Analyser {
	yearPattern := pattern.YearRange(config)
	return New(map[string]pattern.Pattern{yearPattern.Name(): yearPattern})
}

type analyzer struct {
	patterns map[string]pattern.Pattern
}

func (a *analyzer) Analyse(ctx context.Context, source string) messages.ErrorList {
	result := messages.NewErrorList()
	template := Template(ctx)

	if template == "" {
		result.Append(messages.TemplateNotProvided())
		return result
	}
	if source == "" {
		result.Append(messages.NoHeader())
		return result
	}
	templateReader := text.NewReader(template)
	sourceReader := text.NewReader(source)
	for !templateReader.Done() && !sourceReader.Done() {
		if templateReader.Peek() == '{' {
			start := templateReader.Position()
			patternName := readField(templateReader)
			pattern := a.patterns[patternName]
			if pattern == nil {
				result.Append(messages.AnalysisError(start, messages.UnknownPattern(patternName)))
				continue
			}
			result.Append(pattern.Verify(sourceReader).Errors()...)
		}
		if templateReader.Peek() != sourceReader.Peek() {
			result.Append(readDiff(sourceReader, templateReader))
		}
		templateReader.Next()
		sourceReader.Next()
	}

	if !templateReader.Done() {
		result.Append(messages.AnalysisError(sourceReader.Position(), messages.Missed(templateReader.Finish())))
	}

	if !sourceReader.Done() {
		result.Append(messages.AnalysisError(templateReader.Position(), messages.NotExpected(sourceReader.Finish())))
	}

	return result
}

func readDiff(actual, expected text.Reader) error {
	start := actual.Position()
	sb1 := strings.Builder{}
	sb2 := strings.Builder{}
	for !actual.Done() && !expected.Done() {
		r1 := actual.Next()
		r2 := expected.Next()
		if r1 != r2 {
			sb1.WriteRune(r1)
			sb2.WriteRune(r2)
		}
	}
	if actual.Done() && !expected.Done() {
		sb2.WriteString(expected.Finish())
	}
	if !actual.Done() && expected.Done() {
		sb1.WriteString(actual.Finish())
	}
	return messages.AnalysisError(start, messages.Diff(sb1.String(), sb2.String()))
}

func readField(reader text.Reader) string {
	_ = reader.Next()
	defer reader.Next()
	return reader.ReadWhile(func(r rune) bool {
		return r != '}'
	})
}
