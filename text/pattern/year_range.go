package pattern

import (
	"strconv"
	"unicode"

	"github.com/denis-tingajkin/go-header/models"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/text"
)

func readNum(r text.Reader) (int, error) {
	start := r.Position()
	number := r.ReadWhile(unicode.IsDigit)
	result, err := strconv.Atoi(number)
	if err != nil {
		return -1, messages.AnalysisError(start, err)
	}
	return result, nil
}

func YearRange(config models.ReadOnlyConfiguration) Pattern {
	return NewPatternFunc("year range",
		func(r text.Reader) messages.ErrorList {
			start := r.Position()
			result := messages.NewErrorList()
			num, err := readNum(r)
			if err != nil {
				result.Append(err)
				return result
			}
			if num == config.Year() {
				return result
			}
			if num > config.Year() || r.Peek() != '-' {
				result.Append(messages.AnalysisError(start, messages.WrongYear()))
				return result
			}
			r.Next()
			start = r.Position()
			num, err = readNum(r)
			if err != nil {
				result.Append(err)
				return result
			}
			if num == config.Year() {
				return result
			}
			result.Append(messages.AnalysisError(start, messages.WrongYear()))
			return result
		})
}
