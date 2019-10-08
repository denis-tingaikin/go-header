package provider

import (
	"log"

	"github.com/denis-tingajkin/go-header/models"
	"github.com/denis-tingajkin/go-header/provider/git"
	"github.com/denis-tingajkin/go-header/utils"
)

type gitSources struct {
	*git.Git
	config models.ReadOnlyConfiguration
}

//NewGitSources creates git sources provider. Expects valid coniguration.
func NewGitSources(config models.ReadOnlyConfiguration) Sources {
	sources := &gitSources{
		Git:    git.New(config.ProjectDir()),
		config: config,
	}
	return sources
}

func (g *gitSources) Get() []*models.Source {
	var files []string
	if g.config.GoProject() {
		files = utils.GoProjectFiles(g.config.ProjectDir())
	} else {
		files = utils.Files(g.config.ProjectDir())
	}
	if len(files) == 0 {
		return nil
	}
	result := make([]*models.Source, len(files))
	utils.SplitWork(func(index int) {
		file := files[index]
		author := g.Author(file)
		source := &models.Source{
			Path:   file,
			Author: author,
		}
		result[index] = source
	}, g.config.GoroutineCount(), len(files))

	log.Printf("Sources: created %v of %v", len(files), len(result))
	return result
}
