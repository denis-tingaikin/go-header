package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

//DisableLogging disables logging into project
func DisableLogging() {
	log.SetOutput(ioutil.Discard)
}

//MakeFirstLetterUpercase makes first latter of string uppercase
func MakeFirstLetterUpercase(s string) string {
	if len(s) == 0 {
		return s
	}
	builder := strings.Builder{}
	_, _ = builder.WriteRune(unicode.ToUpper(rune(s[0])))
	_, _ = builder.WriteString(s[1:len(s)])
	return builder.String()
}

//IsSuitableGoFile returns true if file has suffix .go and path contains vendor folder
func IsSuitableGoFile(path string) bool {
	return !strings.Contains(path, "vendor\\") && strings.HasSuffix(path, ".go")
}

func AddTabPerLine(s string) string {
	lines := strings.Split(s, "\n")
	sb := strings.Builder{}
	length := len(lines)
	for i := 0; i < length; i++ {
		_, _ = sb.WriteRune('\t')
		_, _ = sb.WriteString(lines[i])
		if i+1 < length {
			_, _ = sb.WriteRune('\n')
		}
	}
	return sb.String()
}

//AllFiles returns all files in dir
func AllFiles(dir string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
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

//SplitWork splits work
func SplitWork(work func(int), splitCount, totalWorkCount int) {
	if work == nil {
		panic("work is nil")
	}
	indexCh := make(chan int)
	go func() {
		for i := 0; i < totalWorkCount; i++ {
			indexCh <- i
		}
		close(indexCh)
	}()
	wg := sync.WaitGroup{}
	wg.Add(splitCount)
	for i := 0; i < splitCount; i++ {
		go func() {
			for index := range indexCh {
				work(index)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
