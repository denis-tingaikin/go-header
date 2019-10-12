package main

import (
	"github.com/denis-tingajkin/go-header/models"
	"github.com/denis-tingajkin/go-header/provider"
)

func main() {
	config := &models.Configuration{
		Rules: []models.Rule{
			models.Rule{
				Template: "rofl",
			},
		},
	}
	result := config.Validate()
	if !result.Empty() {
		println(result.String())
		return
	}
	p := provider.NewGitSources(models.AsReadonly(config))
	m := map[string]bool{}
	for _, s := range p.Get() {
		m[s.Header()] = true
	}
	for k := range m {
		println(k)
		println()
	}
}
