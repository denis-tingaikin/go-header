package models

import (
	"os"
	"runtime"

	"github.com/denis-tingajkin/go-header/messages"
)

//Configuration is main configuration of go-header linter
type Configuration struct {
	//Year means current year for {YEAR} pattern
	Year int `yaml:"year"`
	//GoroutineCount is a count of goroutines for async work
	GoroutineCount int `yaml:"goroutine-count"`
	//ProjectDir is path to scanning project
	ProjectDir string `yaml:"project-dir"`
	//ShowAllErrors means print only the first error of finding errors
	ShowAllErrors bool `yaml:"show-all-errors"`
	//Rules means rules for file matching
	Rules []Rule `yaml:"rules"`
	//CopyrigtHolders means copyright holder for patter {copyright holder}. If empty means any copyright holder.
	CopyrigtHolders []string `yaml:"copyright-holders"`
	//CustomPatterns adds user's patterns
	CustomPatterns []CustomPattern `yaml:"custom-patterns"`
	//Scope provides scope for linting
	Scope Scope `yaml:"scope"`
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
			errList.Append(compileResult.Errors()...)
		}
		if err := rule.loadTemplate(); err != nil {
			errList.Append(err)
		}
	}
}
