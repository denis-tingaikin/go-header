package models

//CustomPattern represents user's patter
type CustomPattern struct {
	//Name means name of pattern
	Name string `yaml:"name"`
	//Pattern represnts source of patter
	Pattern string `yaml:"pattern"`
	//AllowMultiple means pattern can be repeated one or more times.
	AllowMultiple bool `yaml:"allow_multiple"`
}
