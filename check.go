package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/denis-tingajkin/go-header/utils"

	"github.com/denis-tingajkin/go-header/models"
	"github.com/denis-tingajkin/go-header/provider"
	"github.com/denis-tingajkin/go-header/text/analysis"
)

func doCheck(config *models.Configuration) {
	start := time.Now()
	if validationResult := config.Validate(); !validationResult.Empty() {
		fmt.Fprintln(os.Stderr, validationResult.String())
		os.Exit(1)
	}
	pass := true
	readOnlyConfig := models.AsReadonly(config)
	analyser := analysis.NewFromConfig(readOnlyConfig)
	sources := provider.NewGitSources(readOnlyConfig).Get()
	red := color.New(color.FgRed).SprintFunc()
	utils.SplitWork(func(index int) {
		source := sources[index]
		rule := config.FindRule(source)
		if rule == nil {
			log.Printf("can not find rule for source: %v", source)
			return
		}
		ctx := analysis.WithTemplate(context.Background(), rule.Template)
		result := analyser.Analyse(ctx, source.Header())
		if !result.Empty() {
			pass = false
		}
		fmt.Printf("%v\n%v\n\n", source.Path, red(result.String()))
	}, readOnlyConfig.GoroutineCount(), len(sources))
	fmt.Printf("Elapsed: %v\n", time.Now().Sub(start))
	if !pass {
		os.Exit(1)
	}
}
