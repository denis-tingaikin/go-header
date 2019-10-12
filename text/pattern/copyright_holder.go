package pattern

import (
	"strings"

	"github.com/denis-tingajkin/go-header/messages"
	"github.com/denis-tingajkin/go-header/models"
	"github.com/denis-tingajkin/go-header/text"
)

func CopyrightHolder(config models.ReadOnlyConfiguration) Pattern {
	holders := config.CopyrightHolders()
	holdersMap := map[string]bool{}
	used := map[string]bool{}
	for _, holder := range holders {
		key := strings.ToLower(holder)
		holdersMap[key] = true
	}
	containsHolder := func(h string) bool {
		return (holdersMap[h] || (len(holders) == 0 && h != ""))
	}
	return NewPatternFunc("copyright holder", func(r text.Reader) messages.ErrorList {
		start := r.Position()
		result := messages.NewErrorList()
		holder := r.ReadWhile(func(r rune) bool {
			return r != '\n'
		})

		if !containsHolder(strings.ToLower(holder)) {
			result.Append(messages.UnknownCopyrightHolder(start, holder, holders...))
		} else {
			if used[holder] {
				result.Append(messages.CopyrightHolderAlreadyInUse(holder))
			}
			used[holder] = true
		}
		return result
	})
}
