package models

//CustomPattern represents user's patter
type CustomPattern struct {
	//Name means name of pattern
	Name string `yaml:"name"`
	//Pattern is source of custom pattern
	Pattern string `yaml:"pattern"`
	//Separator uses only for multi custom patterns. Means string that splits list of custom patterns
	Separator string `yaml:"separator"`
}

func (c *CustomPattern) AllowMultiple() bool {
	return c.Separator != ""
}
