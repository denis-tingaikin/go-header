// Copyright (c) 2020-2025 Denis Tingaikin
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goheader

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"golang.org/x/tools/go/analysis"
)

type CommentStyleType int

const (
	DoubleSlash CommentStyleType = iota
	MultiLine
	MultiLineStar
)

const iso = "2006-01-02 15:04:05 -0700"

func modTime(path string) (time.Time, error) {
	diff, err := exec.Command("git", "diff", path).CombinedOutput()
	if err == nil && len(diff) == 0 {
		line, err := exec.Command("git", "log", "-1", "--pretty=format:%cd", "--date=iso", "--", path).CombinedOutput()
		if err == nil {
			return time.Parse(iso, string(line))
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

type Analyzer struct {
	Settings *Settings
}

func New(settings *Settings) *analysis.Analyzer {
	analyzer := Analyzer{Settings: settings}

	return &analysis.Analyzer{
		Doc:              "Check file license header",
		URL:              "https://github.com/denis-tingaikin/go-header",
		Name:             "goheader",
		RunDespiteErrors: true,
		Run:              analyzer.Run,
	}
}

type Result struct {
	Fix        string
	End, Start token.Pos
}

func (a *Analyzer) skipCodeGen(file *ast.File) ([]*ast.CommentGroup, []*ast.Comment) {
	var comments = file.Comments
	var list []*ast.Comment
	if len(comments) > 0 {
		list = comments[0].List
	}
	if len(comments) > 0 && strings.Contains(comments[0].Text(), "DO NOT EDIT") {
		comments = comments[1:]
		list = comments[0].List
		if len(list) > 0 && strings.HasSuffix(list[0].Text, "//line:") {
			list = list[1:]
		}
	}

	for len(list) > 0 {
		if a.isDirective(list[0].Text) {
			list = list[1:]
			if len(list) == 0 {
				comments = comments[1:]
				if len(comments) > 0 {
					list = comments[0].List
				}
			}
			continue
		}
		break
	}

	return comments, list
}

func (a *Analyzer) Run(pass *analysis.Pass) (any, error) {
	jobCh := make(chan *ast.File, len(pass.Files))

	for _, f := range pass.Files {
		file := f
		jobCh <- file
	}
	close(jobCh)

	var wg sync.WaitGroup

	for range a.Settings.Parallel {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for file := range jobCh {
				filename := pass.Fset.PositionFor(file.Pos(), false).Filename
				if !strings.HasSuffix(filename, ".go") {
					continue
				}

				diag, err := a.Analyze(filename, file)
				if err != nil {
					// TODO handle the error.
					return
				}

				if diag == nil {
					continue
				}

				var line = 1
				if ast.IsGenerated(file) {
					line = 4
				}

				fileToken := pass.Fset.File(file.Pos())

				start := fileToken.LineStart(line)
				endLine := fileToken.Line(diag.End-diag.Pos+start) + 1

				var end token.Pos
				if endLine < fileToken.LineCount() {
					end = fileToken.LineStart(endLine)
				} else {
					end = start
				}

				diag.Pos = start
				diag.End = end

				if len(diag.SuggestedFixes) > 0 && len(diag.SuggestedFixes[0].TextEdits) > 0 {
					diag.SuggestedFixes[0].TextEdits[0].Pos = start
					diag.SuggestedFixes[0].TextEdits[0].End = end
				}

				pass.Report(*diag)
			}
		}()
	}

	wg.Wait()

	return nil, nil
}

func (a *Analyzer) Analyze(path string, file *ast.File) (*analysis.Diagnostic, error) {
	if a.Settings.Template == "" {
		return nil, nil
	}

	var header string
	var style CommentStyleType

	var comments, list = a.skipCodeGen(file)

	result := &analysis.Diagnostic{}

	if len(comments) > 0 && comments[0].Pos() < file.Package {
		if strings.HasPrefix(list[0].Text, "/*") {
			result.Pos = list[0].Pos()
			result.End = list[0].End()

			header = (&ast.CommentGroup{List: []*ast.Comment{list[0]}}).Text()
			style = MultiLine

			if handledHeader, ok := handleStarBlock(header); ok {
				header = handledHeader
				style = MultiLineStar
			}

		} else {
			result.Pos = comments[0].Pos()
			result.End = comments[0].Pos()

			header = comments[0].Text()
			style = DoubleSlash
		}
	}

	vars, err := a.getPerTargetValues(path)
	if err != nil {
		return nil, err
	}

	header = strings.TrimSpace(header)

	if header == "" {
		text, err := a.generateFix(style, vars)
		if err != nil {
			return nil, err
		}

		result.Message = "missed copyright header"
		result.SuggestedFixes = append(result.SuggestedFixes, analysis.SuggestedFix{
			TextEdits: []analysis.TextEdit{{
				NewText: []byte(text),
			}},
		})

		return result, nil
	}

	templateRaw := a.quoteMeta(a.Settings.Template)

	tmpl, err := template.New("header").Delims(a.Settings.LeftDelim, a.Settings.RightDelim).Parse(templateRaw)
	if err != nil {
		return nil, err
	}

	headerTemplateBuffer := new(bytes.Buffer)

	err = tmpl.Execute(headerTemplateBuffer, vars)
	if err != nil {
		return nil, err
	}

	exp, err := regexp.Compile(headerTemplateBuffer.String())
	if err != nil {
		return nil, err
	}

	if !exp.MatchString(header) {
		text, _ := a.generateFix(style, vars)

		result.Message = "template doesn't match"
		if text != "" {
			result.SuggestedFixes = append(result.SuggestedFixes, analysis.SuggestedFix{
				TextEdits: []analysis.TextEdit{{
					NewText: []byte(text),
				}},
			})
		}

		return result, nil
	}

	return nil, nil
}

var directiveRegexp = regexp.MustCompile(`([a-z0-9]+:[a-z0-9])|(\+build)`)

func (a *Analyzer) isDirective(comment string) bool {
	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimPrefix(comment, "/*")
	comment = strings.TrimSpace(comment)

	comment = strings.Split(comment, " ")[0]

	if strings.HasPrefix(comment, "line") || strings.HasPrefix(comment, "extern") || strings.HasPrefix(comment, "export") {
		return true
	}

	return directiveRegexp.Match([]byte(comment))
}

func (a *Analyzer) generateFix(style CommentStyleType, vals map[string]Value) (string, error) {
	// TODO: add values for quick fixes in config
	vals["YEAR_RANGE"] = vals["YEAR"]
	vals["MOD_YEAR_RANGE"] = vals["YEAR"]

	for _, v := range vals {
		if _, ok := v.(*RegexpValue); ok {
			return "", errors.New("fixes are not supported for regexp values. See more details https://github.com/denis-tingaikin/go-header/issues/52")
		}
		_ = v.Calculate(vals)
	}

	fixTemplate, err := template.New("fix").Parse(a.Settings.Template)
	if err != nil {
		return "", err
	}

	fixOut := new(bytes.Buffer)
	err = fixTemplate.Execute(fixOut, vals)
	if err != nil {
		return "", err
	}

	resSplit := strings.Split(fixOut.String(), "\n")

	for i := range resSplit {
		switch style {
		case DoubleSlash:
			resSplit[i] = "// " + resSplit[i]
		case MultiLineStar:
			resSplit[i] = " * " + resSplit[i]
		case MultiLine:
			continue
		}
	}

	switch style {
	case MultiLineStar:
		resSplit = append([]string{"/*"}, resSplit...)
		resSplit = append(resSplit, " */")
	case MultiLine:
		resSplit = append([]string{"/*"}, resSplit...)
		resSplit = append(resSplit, "*/")
	}

	return strings.Join(resSplit, "\n") + "\n", nil
}

func (a *Analyzer) getPerTargetValues(path string) (map[string]Value, error) {
	var res = make(map[string]Value)

	for k, v := range a.Settings.Values {
		res[k] = v
	}

	res["MOD_YEAR"] = a.Settings.Values["YEAR"]
	res["MOD_YEAR_RANGE"] = a.Settings.Values["YEAR_RANGE"]
	if t, err := modTime(path); err == nil {
		res["MOD_YEAR"] = &ConstValue{RawValue: fmt.Sprint(t.Year())}
		res["MOD_YEAR_RANGE"] = &RegexpValue{RawValue: `((20\d\d\-{{.MOD_YEAR}})|({{.MOD_YEAR}}))`}
	}

	for _, v := range res {
		if err := v.Calculate(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// TODO: Do not vibe code
func (a *Analyzer) quoteMeta(text string) string {
	var result strings.Builder
	var i int

	n := len(text)
	for i < n {
		// Check for template placeholder start
		if i+3 < n && text[i] == a.Settings.LeftDelim[0] && text[i+1] == a.Settings.LeftDelim[1] {
			// Find the end of the placeholder
			end := i + 2
			for end < n && !(text[end] == a.Settings.RightDelim[0] && end+1 < n && text[end+1] == a.Settings.RightDelim[1]) {
				end++
			}
			if end+1 < n {
				// Append the entire placeholder without escaping
				result.WriteString(text[i : end+2])
				i = end + 2
				continue
			}
		}

		// Escape regular expression metacharacters for non-template parts
		c := text[i]
		if strings.ContainsAny(string(c), `\.+*?()|[]{}^$`) {
			result.WriteByte('\\')
		}
		result.WriteByte(c)

		i++
	}

	return result.String()
}

func handleStarBlock(header string) (string, bool) {
	var handled = false
	return trimEachLine(header, func(s string) string {
		var trimmed = strings.TrimSpace(s)
		if !strings.HasPrefix(trimmed, "*") {
			return s
		}
		if v, ok := strings.CutPrefix(trimmed, "* "); ok {
			handled = true
			return v
		} else {
			var res, _ = strings.CutPrefix(trimmed, "*")
			return res
		}
	}), handled
}

func trimEachLine(input string, trimFunc func(string) string) string {
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		lines[i] = trimFunc(line)
	}

	return strings.Join(lines, "\n")
}
