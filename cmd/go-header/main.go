package main

import (
	"fmt"
	goheader "github.com/denis-tingajkin/go-header"
	"github.com/denis-tingajkin/go-header/version"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"go/parser"
	"go/token"
	"os"
)

const configPath = ".go-header.yml"

type issue struct {
	goheader.Issue
	filePath string
}

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		logrus.Fatal("Paths has not passed")
	}
	if len(paths) == 1 {
		if paths[0] == "version" {
			fmt.Println(version.Value())
			return
		}
	}
	c := &goheader.Configuration{}
	if err := c.Parse(configPath); err != nil {
		logrus.Fatal(err.Error())
	}
	v, err := c.GetValues()
	if err != nil {
		logrus.Fatalf("Can not get values: %v", err.Error())
	}
	t, err := c.GetTemplate()
	if err != nil {
		logrus.Fatalf("Can not get template: %v", err.Error())
	}
	a := goheader.New(goheader.WithValues(v), goheader.WithTemplate(t))
	s := token.NewFileSet()
	var issues []*issue
	for _, p := range paths {
		f, err := parser.ParseFile(s, p, nil, parser.ParseComments)
		if err != nil {
			logrus.Fatalf("File %v can not be parsed due compilation errors %v", p, err.Error())
		}
		i := a.Analyze(f)
		if i != nil {
			issues = append(issues, &issue{
				Issue:    i,
				filePath: p,
			})
		}
	}
	if len(issues) > 0 {
		red := color.New(color.FgRed).SprintFunc()
		for _, issue := range issues {
			fmt.Printf("%v:%v\n%v\n", issue.filePath, issue.Location(), red(issue.Message()))
		}
		os.Exit(-1)
	}
}
