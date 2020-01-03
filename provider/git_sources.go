package provider

import (
	"log"
	"path"

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
	scope := g.config.Scope()
	if scope.Policy == models.DiffOnlyPolicy {
		files = g.Git.DiffFiles(scope.MasterBranch)
	} else if scope.Policy == models.OnlyNewFilesPolicy {
		files = g.Git.OnlyNewFiles(scope.MasterBranch)
	} else {
		files = utils.AllFiles(g.config.ProjectDir())
	}
	if len(files) == 0 {
		return nil
	}
	result := make([]*models.Source, len(files))
	utils.SplitWork(func(index int) {
		file := files[index]
		author := g.Author(file)
		filePath := file
		if scope.Policy != models.NonePolicy {
			filePath = path.Join(g.config.ProjectDir(), filePath)
		}
		source := &models.Source{
			Path:   filePath,
			Author: author,
		}
		result[index] = source
	}, g.config.GoroutineCount(), len(files))

	log.Printf("Sources: created %v of %v", len(files), len(result))
	return result
}
