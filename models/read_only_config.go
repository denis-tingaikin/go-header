package models

type ReadOnlyConfiguration interface {
	Year() int
	GoroutineCount() int
	ProjectDir() string
	GoProject() bool
	ShowOnlyFirstError() bool
	CopyrightHolders() []string
	Scope() Scope
	CustomPatterns() []CustomPattern
}

type readOnlyConfiguration struct {
	config *Configuration
}

func (r *readOnlyConfiguration) Year() int {
	return r.config.Year
}
func (r *readOnlyConfiguration) GoroutineCount() int {
	return r.config.GoroutineCount
}
func (r *readOnlyConfiguration) ProjectDir() string {
	return r.config.ProjectDir
}

func (r *readOnlyConfiguration) GoProject() bool {
	return r.config.GoProject
}

func (r *readOnlyConfiguration) ShowOnlyFirstError() bool {
	return r.config.ShowOnlyFirstError
}

func (r *readOnlyConfiguration) CopyrightHolders() []string {
	result := make([]string, len(r.config.CopyrigtHolders))
	copy(result, r.config.CopyrigtHolders)
	return result
}

func (r *readOnlyConfiguration) Scope() Scope {
	return r.config.Scope
}

func (r *readOnlyConfiguration) CustomPatterns() []CustomPattern {
	result := make([]CustomPattern, len(r.config.CustomPatterns))
	copy(result, r.config.CustomPatterns)
	return r.config.CustomPatterns
}

func AsReadonly(config *Configuration) ReadOnlyConfiguration {
	return &readOnlyConfiguration{config: config}
}
