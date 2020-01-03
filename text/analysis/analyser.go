package analysis

import (
	"context"
	"fmt"
	"strings"

	"github.com/denis-tingajkin/go-header/models"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/text"
)

type Analyser interface {
	Analyse(context.Context, string) messages.ErrorList
}

func NewFromConfig(config models.ReadOnlyConfiguration) Analyser {
	customPatterns := map[string]*models.CustomPattern{}
	configPatterns := config.CustomPatterns()
	for i := range configPatterns {
		pattern := &configPatterns[i]
		key := strings.ToLower(pattern.Name)
		customPatterns[key] = pattern
	}
	return &analyzer{
		customPatterns: customPatterns,
		config:         config,
	}
}

type analyzer struct {
	customPatterns map[string]*models.CustomPattern
	config         models.ReadOnlyConfiguration
}

func (a *analyzer) Analyse(ctx context.Context, source string) messages.ErrorList {
	template := Template(ctx)
	if template == "" {
		return messages.NewErrorList(messages.TemplateNotProvided())
	}
	if source == "" {
		return messages.NewErrorList(messages.NoHeader())
	}
	templateReader := text.NewReader(template)
	sourceReader := text.NewReader(source)

	result := a.analyzeReaders(WithPatterns(ctx, a.config), sourceReader, templateReader)

	if !templateReader.Done() {
		result.Append(messages.AnalysisError(sourceReader.Location(), messages.Missed(templateReader.Finish())))
	}

	if !sourceReader.Done() {
		result.Append(messages.AnalysisError(templateReader.Location(), messages.NotExpected(sourceReader.Finish())))
	}
	return result
}

func (a *analyzer) analyzeReaders(ctx context.Context, sourceReader, templateReader text.Reader) messages.ErrorList {
	result := messages.NewErrorList()
	potentialErrors := messages.NewErrorList()
	for !templateReader.Done() && !sourceReader.Done() {
		if templateReader.Peek() == '{' {
			start := templateReader.Location()
			patternName := strings.ToLower(a.readField(templateReader))
			pattern := FindPattern(ctx, patternName)
			customPattern := a.customPatterns[patternName]
			if err := a.checkLoop(ctx, patternName); err != nil {
				result.Append(err)
				return result
			}
			if pattern != nil {
				result.Append(pattern.Verify(sourceReader).Errors()...)
				continue
			}
			if customPattern != nil {
				ctx = Visit(ctx, patternName)
				errs := a.analyzeReaders(ctx, sourceReader, text.NewReader(customPattern.Pattern))
				if !errs.Empty() {
					result.Append(errs.Errors()...)
				} else if customPattern.AllowMultiple() {
					potentialErrors.Append(a.readMultiplePattern(ctx, sourceReader, customPattern).Errors()...)
				}
				ctx = Leave(ctx, patternName)
				continue
			}
			result.Append(messages.AnalysisError(start, messages.UnknownPattern(patternName)))
		}
		if templateReader.Peek() != sourceReader.Peek() {
			result.Append(a.readDiff(sourceReader, templateReader))
		}
		templateReader.Next()
		sourceReader.Next()
	}
	if !result.Empty() && !potentialErrors.Empty() {
		if a.config.ShowAllErrors() {
			return messages.NewErrorList(messages.Ambiguous(result, potentialErrors))
		}
		return messages.NewErrorList(messages.Ambiguous(messages.NewErrorList(result.Errors()[0]), messages.NewErrorList(potentialErrors.Errors()[0])))
	}
	return result
}

func (a *analyzer) checkLoop(ctx context.Context, n string) error {
	if IsVisited(ctx, n) {
		return messages.DetectedInfiniteRecursiveEntry(append(VisitedList(ctx), n)...)
	}
	return nil
}

func (a *analyzer) readMultiplePattern(ctx context.Context, source text.Reader, pattern *models.CustomPattern) messages.ErrorList {
	for !source.Done() {
		pop := source.Position()
		p := source.ReadWhile(text.LengthNotEqual(len(pattern.Separator)))
		if p != pattern.Separator {
			source.SetPosition(pop)
			return messages.NewErrorList(messages.Missed(fmt.Sprintf("separator: \"%v\"", pattern.Name)))
		}
		errs := a.analyzeReaders(ctx, source, text.NewReader(pattern.Pattern))
		if !errs.Empty() {
			source.SetPosition(pop)
			return errs
		}
	}
	return messages.NewErrorList()
}

func (a *analyzer) readDiff(actual, expected text.Reader) error {
	start := actual.Location()
	sb1 := strings.Builder{}
	sb2 := strings.Builder{}
	for !actual.Done() && !expected.Done() {
		r1 := actual.Next()
		r2 := expected.Next()
		if r1 == r2 {
			break
		}
		sb1.WriteRune(r1)
		sb2.WriteRune(r2)
	}
	if actual.Done() && !expected.Done() {
		sb2.WriteString(expected.Finish())
	}
	if !actual.Done() && expected.Done() {
		sb1.WriteString(actual.Finish())
	}
	return messages.AnalysisError(start, messages.Diff(sb1.String(), sb2.String()))
}

func (a *analyzer) readField(reader text.Reader) string {
	_ = reader.Next()
	defer reader.Next()
	return reader.ReadWhile(func(r rune) bool {
		return r != '}'
	})
}
