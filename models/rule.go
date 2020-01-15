package models

import (
	"io/ioutil"
	"path"
	"regexp"

	"github.com/denis-tingajkin/go-header/messages"
)

//Rule means rule for matching files
type Rule struct {
	//Template means license header for files
	Template string `yaml:"template"`
	//TemplatePath means license header for files located to specific folder
	TemplatePath string `yaml:"template-path"`
	//PathMatcher means regex for file path
	Paths []string `yaml:"paths"`
	//AuthorMatcher means author regex for authors
	Authors []string `yaml:"authors"`
	//ExcludePathMatcher means regex pattern to exclude files
	ExcludePaths []string `yaml:"exclude-paths"`
	authors      []*regexp.Regexp
	paths        []*regexp.Regexp
	excludePaths []*regexp.Regexp
}

func (r *Rule) loadTemplate(projectDir string) error {
	if r.Template == "" && r.TemplatePath != "" {
		bytes, err := ioutil.ReadFile(path.Join(projectDir, r.TemplatePath))
		if err != nil {
			return messages.CanNotLoadTemplateFromFile(err)
		}
		r.Template = string(bytes)
	}
	if r.Template == "" {
		return messages.TemplateNotProvided()
	}
	return nil
}

func (r *Rule) Compile() messages.ErrorList {
	result := messages.NewErrorList()
	var err error
	if r.authors, err = compileRegularExpressions(r.Authors); err != nil {
		result.Append(err)
	}
	if r.paths, err = compileRegularExpressions(r.Paths); err != nil {
		result.Append(err)
	}
	if r.excludePaths, err = compileRegularExpressions(r.ExcludePaths); err != nil {
		result.Append(err)
	}
	return result
}

func (r Rule) Match(s *Source) bool {
	if len(r.excludePaths) != 0 && anyMatch(r.excludePaths, s.Path) {
		return false
	}
	if len(r.paths) != 0 && !anyMatch(r.paths, s.Path) {
		return false
	}
	return len(r.authors) == 0 || anyMatch(r.authors, s.Author)
}

func compileRegularExpressions(regexpSources []string) ([]*regexp.Regexp, error) {
	var result []*regexp.Regexp
	for _, s := range regexpSources {
		var r *regexp.Regexp
		var err error
		if r, err = regexp.Compile(s); err != nil {
			return nil, messages.CantProcessField(s, err)
		}
		result = append(result, r)
	}
	return result, nil
}

func anyMatch(regexps []*regexp.Regexp, s string) bool {
	for _, r := range regexps {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}
