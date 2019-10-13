package models

import (
	"os"
	"runtime"

	"github.com/denis-tingajkin/go-header/messages"
)

type Configuration struct {
	Year               int             `yaml:"year"`
	GoProject          bool            `yaml:"go-project"`
	GoroutineCount     int             `yaml:"goroutine-count"`
	ProjectDir         string          `yaml:"project_dir"`
	ShowOnlyFirstError bool            `yaml:"show-only-first-error"`
	Rules              []Rule          `yaml:"rules"`
	CopyrigtHolders    []string        `yaml:"copyright-holders"`
	CustomPatterns     []CustomPattern `yaml:"custom-patterns"`
	Scope              Scope           `yaml:"scope"`
}

func (c *Configuration) FindRule(s *Source) *Rule {
	for i := range c.Rules {
		rule := &c.Rules[i]
		if rule.Match(s) {
			return rule
		}
	}
	return nil
}

func (c *Configuration) Validate() messages.ErrorList {
	result := messages.NewErrorList()
	if c.ProjectDir == "" {
		var err error
		c.ProjectDir, err = os.Getwd()
		if err != nil {
			result.Append(err)
		}
	}
	if c.GoroutineCount < 0 {
		result.Append(messages.IncorrectGoroutineCount(c.GoroutineCount))
	}
	if c.GoroutineCount == 0 {
		c.GoroutineCount = runtime.NumCPU()
	}
	if len(c.Rules) == 0 {
		result.Append(messages.NoRules())
		return result
	}
	c.checkRules(result)
	if err := c.Scope.Validate(); err != nil {
		result.Append(err)
	}
	return result
}

func (c *Configuration) checkRules(errList messages.ErrorList) {
	for i := range c.Rules {
		rule := &c.Rules[i]
		if compileResult := rule.Compile(); !compileResult.Empty() {
			errIndex := 0
			if rule.pathMatcher == nil && rule.PathMatcher != "" {
				errList.Append(messages.CantProcessField(rule.PathMatcher, compileResult.Errors()[errIndex]))
				errIndex++
			}
			if rule.authorMatcher == nil && rule.AuthorMatcher != "" {
				errList.Append(messages.CantProcessField(rule.AuthorMatcher, compileResult.Errors()[errIndex]))
				errIndex++
			}
			if rule.excludePathMatcher == nil && rule.ExcludePathMatcher != "" {
				errList.Append(messages.CantProcessField(rule.ExcludePathMatcher, compileResult.Errors()[errIndex]))
				errIndex++
			}
		}
		if err := rule.loadTemplate(); err != nil {
			errList.Append(err)
		}
	}
}
