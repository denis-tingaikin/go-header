package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//DisableLogging disables logging into project
func DisableLogging() {
	log.SetOutput(ioutil.Discard)
}

//GoProjectFiles returns all .go files in dir. Excludes vendor folder.
func GoProjectFiles(dir string) []string {
	filterFunc := func(path string) bool {
		relativePath := path[len(dir):]
		return !strings.HasPrefix(relativePath, `\vendor`) && strings.HasSuffix(relativePath, ".go")
	}
	return files(dir, filterFunc)
}

//Files returns all files into dir
func Files(dir string) []string {
	return files(dir, func(string) bool {
		return true
	})
}

func files(dir string, filterFunc func(string) bool) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filterFunc(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Printf("An error during scanning dir: %v. Error: %v", dir, err.Error())
		return nil
	}
	return files
}
