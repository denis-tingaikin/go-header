# go-header

Go source code linter providing checks for license headers.

# Installation

For installation you can simply use `go get`.

```
go get github.com/denis-tingajkin/go-header
```

# Configuration

For configuring `go-header` linter you simply need to fill the next structures in YAML format.
```
type Configuration struct {
	Year               int             `yaml:"year"`
	GoProject          bool            `yaml:"go-project"`
	GoroutineCount     int             `yaml:"goroutine-count"`
	ProjectDir         string          `yaml:"project-dir"`
	ShowOnlyFirstError bool            `yaml:"show-only-first-error"`
	Rules              []Rule          `yaml:"rules"`
	CopyrigtHolders    []string        `yaml:"copyright-holders"`
	CustomPatterns     []CustomPattern `yaml:"custom-patterns"`
	Scope              Scope           `yaml:"scope"`
}
```
Scope structure
```
//Scope means the scope for go-header linter in project
type Scope struct {
	//Policy means file scoe policy. Can be "none", "diff", "new".
	Policy GitPolicy `yaml:"policy"`
	//MasterBranch master branch for scope. Used only if Policy is not "none".
	MasterBranch string `yaml:"master-branch"`
}

```
Custom patterns structure:
```
//CustomPattern represents user's patter
type CustomPattern struct {
	//Name means name of pattern
	Name string `yaml:"name"`
	//Pattern represnts source of patter
	Pattern string `yaml:"pattern"`
	//AllowMultiple means pattern can be repeated one or more times.
	AllowMultiple bool `yaml:"allow_multiple"`
}
```
Rule structure:
```
//Rule means rule for matching files
type Rule struct {
	//Template means license header for files
	Template string `yaml:"template"`
	//TemplatePath reads header from specific folder 
	TemplatePath string `yaml:"template-path"`
	//PathMatcher means regex for file path
	PathMatcher string `yaml:"path-matcher"`
	//AuthorMatcher means author regex for authors
	AuthorMatcher string `yaml:"author-matcher"`
	//ExcludePathMatcher means regex pattern to exclude files
	ExcludePathMatcher string `yaml:"exclude-path-matcher"`
	authorMatcher      *regexp.Regexp
	pathMatcher        *regexp.Regexp
	excludePathMatcher *regexp.Regexp
}

```

# Flags

`-logging=true` means enable logging. Useful for bug reporting.

`-path=...` means path to go-header .yaml configuraiton.





