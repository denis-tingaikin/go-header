package models

type ReadOnlyConfiguration interface {
	Year() int
	GoroutineCount() int
	ProjectDir() string
	GoProject() bool
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

func AsReadonly(config *Configuration) ReadOnlyConfiguration {
	return &readOnlyConfiguration{config: config}
}
