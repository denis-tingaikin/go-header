package models

type CustomPattern struct {
	Name          string `yaml:"name"`
	Pattern       string `yaml:"pattern"`
	AllowMultiple bool   `yaml:"allow_multiple"`
}
