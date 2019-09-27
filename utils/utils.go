package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//DisableLogging disables logging into project
func DisableLogging() {
	log.SetOutput(ioutil.Discard)
}

//GoProjectFiles returns all .go files in dir. Excludes vendor folder.
func GoProjectFiles(dir string) []string {
	filterFunc := func(path string) bool {
		relativePath := path[len(dir):]
		return !strings.HasPrefix(relativePath, `\vendor`) && strings.HasSuffix(relativePath, ".go") && !strings.HasSuffix(relativePath, "test.go")
	}
	return files(dir, filterFunc)
}

//Files returns all files into dir
func Files(dir string) []string {
	return files(dir, func(string) bool {
		return true
	})
}

//SplitWork splits work
func SplitWork(work func(int), splitCount, totalWorkCount int) {
	if work == nil {
		panic("work is nil")
	}
	wg := sync.WaitGroup{}
	wg.Add(splitCount)

	step := totalWorkCount / splitCount

	for i := 0; i < splitCount; i++ {
		body := func(start, end int) {
			for workIndex := start; workIndex < end; workIndex++ {
				work(workIndex)
			}
			wg.Done()
		}
		if i+1 == splitCount {
			go body(i*step, totalWorkCount)
		} else {
			go body(i*step, (i+1)*step)
		}
	}
	wg.Wait()
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
