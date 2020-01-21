package main

import (
	"flag"
	"fmt"
	"github.com/denis-tingajkin/go-header/version"
	"io/ioutil"
	"os"

	"github.com/denis-tingajkin/go-header/utils"

	"gopkg.in/yaml.v2"

	"github.com/denis-tingajkin/go-header/models"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		println(version.Get())
		return
	}
	pathToFile := flag.String("path", ".go-header.yaml", "provides path to config.yaml file")
	logging := flag.Bool("logging", false, "enables logging in to stdout")
	flag.Parse()
	if !*logging {
		utils.DisableLogging()
	}
	config := new(models.Configuration)
	bytes, err := ioutil.ReadFile(*pathToFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error during read file '%v'. Err: %v\n", *pathToFile, err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error during parse '%v'. Err: %v\n", *pathToFile, err)
		os.Exit(1)
	}
	doCheck(config)
}
