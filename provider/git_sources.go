package provider

import (
	"log"
	"runtime"

	"github.com/go-header/models"
	"github.com/go-header/provider/git"
	"github.com/go-header/utils"
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
	goroutineCount := runtime.NumCPU()
	if g.config.GoroutineCount() != 0 {
		goroutineCount = g.config.GoroutineCount()
	}
	step := len(files) / goroutineCount
	channels := make([]<-chan []*models.Source, goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		if i+1 == goroutineCount {
			channels[i] = g.collectSources(files, i*step, len(files))
		} else {
			channels[i] = g.collectSources(files, i*step, (i+1)*step)
		}
	}
	result := []*models.Source{}
	for _, ch := range channels {
		result = append(result, <-ch...)
	}
	log.Printf("Sources: created %v of %v", len(files), len(result))
	return result
}

func (g *gitSources) collectSources(files []string, start, end int) <-chan []*models.Source {
	ch := make(chan []*models.Source)
	go func() {
		result := []*models.Source{}
		for i := start; i < end; i++ {
			file := files[i]
			author := g.Author(file)
			source := &models.Source{
				Path:   file,
				Author: author,
			}
			result = append(result, source)
		}
		ch <- result
	}()
	return ch
}
