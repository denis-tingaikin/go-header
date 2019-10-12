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

func NewFromConfig(config models.ReadOnlyConfiguration) Analyser {
	yearPattern := pattern.YearRange(config)
	copyrightHolderPattern := pattern.CopyrightHolder(config)
	customPatterns := map[string]*models.CustomPattern{}
	configPatterns := config.CustomPatterns()
	for i := range configPatterns {
		pattern := &configPatterns[i]
		customPatterns[pattern.Name] = pattern
	}
	return &analyzer{
		patterns: map[string]pattern.Pattern{
			yearPattern.Name():            yearPattern,
			copyrightHolderPattern.Name(): copyrightHolderPattern,
		},
		customPatterns: customPatterns,
		visit:          map[string]bool{},
	}
}

type analyzer struct {
	patterns       map[string]pattern.Pattern
	customPatterns map[string]*models.CustomPattern
	visit          map[string]bool
	visited        []string
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

	result := a.analyzeReaders(sourceReader, templateReader)

	if !templateReader.Done() {
		result.Append(messages.AnalysisError(sourceReader.Position(), messages.Missed(templateReader.Finish())))
	}

	if !sourceReader.Done() {
		result.Append(messages.AnalysisError(templateReader.Position(), messages.NotExpected(sourceReader.Finish())))
	}
	return result
}

func (a *analyzer) analyzeReaders(sourceReader, templateReader text.Reader) messages.ErrorList {
	result := messages.NewErrorList()
	potentialErrors := messages.NewErrorList()
	for !templateReader.Done() && !sourceReader.Done() {
		if templateReader.Peek() == '{' {
			start := templateReader.Position()
			patternName := strings.ToLower(a.readField(templateReader))
			pattern := a.patterns[patternName]
			customPattern := a.customPatterns[patternName]
			if err := a.checkLoop(patternName); err != nil {
				result.Append(err)
				return result
			}
			if pattern != nil {
				result.Append(pattern.Verify(sourceReader).Errors()...)
				continue
			}
			if customPattern != nil {
				a.visit[patternName] = true
				a.visited = append(a.visited, patternName)
				errs := a.analyzeReaders(sourceReader, text.NewReader(customPattern.Pattern))
				if !errs.Empty() {
					result.Append(errs.Errors()...)
				} else if customPattern.AllowMultiple {
					potentialErrors.Append(a.readMultiplePattern(sourceReader, customPattern).Errors()...)
				}
				a.visit[patternName] = false
				a.visited = a.visited[0 : len(a.visited)-1]
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
		return messages.NewErrorList(messages.Ambiguous(result, potentialErrors))
	}
	return result
}

func (a *analyzer) checkLoop(n string) error {
	if a.visit[n] {
		return messages.DetectedInfiniteRecursiveEntry(append(a.visited, n)...)
	}
	return nil
}

func (a *analyzer) readMultiplePattern(source text.Reader, patter *models.CustomPattern) messages.ErrorList {
	for !source.Done() {
		pop := source.Position()
		errs := a.analyzeReaders(source, text.NewReader(patter.Pattern))
		if !errs.Empty() {
			source.SetPosition(pop)
			return errs
		}
	}
	return messages.NewErrorList()
}

func (a *analyzer) readDiff(actual, expected text.Reader) error {
	start := actual.Position()
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
