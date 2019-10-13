package pattern

import (
	"strconv"
	"unicode"

	"github.com/denis-tingajkin/go-header/models"

	"github.com/denis-tingajkin/go-header/messages"

	"github.com/denis-tingajkin/go-header/text"
)

func readNum(r text.Reader) (int, error) {
	start := r.Location()
	number := r.ReadWhile(unicode.IsDigit)
	if number == "" {
		return -1, messages.AnalysisError(start, messages.CatNotParseAsYear())
	}
	result, err := strconv.Atoi(number)
	return result, err
}

func YearRange(config models.ReadOnlyConfiguration) Pattern {
	return NewPatternFunc("year",
		func(r text.Reader) messages.ErrorList {
			start := r.Location()
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
			start = r.Location()
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
