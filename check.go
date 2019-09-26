package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-header/messages"
	"github.com/go-header/models"
	"github.com/go-header/provider"
	"github.com/go-header/text/analysis"
)

func doCheck(config *models.Configuration) {
	result := config.Validate()
	if !result.Empty() {
		fmt.Fprintln(os.Stderr, result.String())
		os.Exit(1)
	}
	pass := true
	readOnlyConfig := models.AsReadonly(config)
	a := analysis.NewFromConfig(readOnlyConfig)
	sources := provider.NewGitSources(readOnlyConfig).Get()
	step := len(sources) / readOnlyConfig.GoroutineCount()
	wg := sync.WaitGroup{}
	wg.Add(readOnlyConfig.GoroutineCount())
	for i := 0; i < readOnlyConfig.GoroutineCount(); i++ {
		go func(index int) {
			results := analyseSources(a, config, sources[index*step:(index+1)*step])
			for _, result := range results {
				if !result.errList.Empty() {
					pass = false
				}
				fmt.Println(result)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	if !pass {
		os.Exit(1)
	}
}

func analyseSources(a analysis.Analyser, conf *models.Configuration, sources []*models.Source) []*analysisResult {
	result := []*analysisResult{}
	for i := range sources {
		source := sources[i]
		rule := conf.FindRule(source)
		if rule == nil {
			log.Printf("can not find rule for source: %v", source)
			continue
		}
		ctx := analysis.WithTemplate(context.Background(), rule.Template)
		result = append(result, &analysisResult{
			filePath: source.Path,
			errList:  a.Analyse(ctx, source.Header()),
		})
	}
	return result
}

type analysisResult struct {
	errList  messages.ErrorList
	filePath string
}

func (a *analysisResult) String() string {
	return fmt.Sprintf("%v:\n%v\n", a.filePath, a.errList.String())
}
