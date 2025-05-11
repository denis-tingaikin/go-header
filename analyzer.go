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
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	values                            map[string]Value
	template, delimsLeft, delimsRight string
}

func New(opts ...Option) *Analyzer {
	var a = Analyzer{
		delimsLeft:  "{{",
		delimsRight: "}}",
	}
	for _, opt := range opts {
		opt.apply(&a)
	}
	return &a
}

type Result struct {
	Err        error
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

	return comments, list
}

func (a *Analyzer) Analyze(path string, file *ast.File) (result analysis.Diagnostic) {
	header := ""

	if a.template == "" {
		return result
	}

	var style CommentStyleType

	var comments, list = a.skipCodeGen(file)

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
			style = DoubleSlash
			header = comments[0].Text()
			result.Pos = comments[0].Pos()
			result.End = comments[0].Pos()
		}
	}
	header = strings.TrimSpace(header)

	vars, err := a.getPerTargetValues(path, file)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	if header == "" {
		result.Message = "missed copyright header"
		result.SuggestedFixes = append(result.SuggestedFixes, analysis.SuggestedFix{
			TextEdits: []analysis.TextEdit{
				{
					NewText: []byte(a.generateFix(style, vars)),
				},
			},
		})
		return result
	}

	templateRaw := a.quoteMeta(a.template)

	template, err := template.New("header").Delims(a.delimsLeft, a.delimsRight).Parse(templateRaw)
	if err != nil {
		result.Message = err.Error()
		return result
	}

	headerTemplateBuffer := new(bytes.Buffer)

	if err := template.Execute(headerTemplateBuffer, vars); err != nil {
		result.Message = err.Error()
		return result
	}

	headerTemplate := headerTemplateBuffer.String()

	r, err := regexp.Compile(headerTemplate)

	if err != nil {
		result.Message = err.Error()
		return result
	}

	if !r.MatchString(header) {
		result.Message = "template doesn't match"
		result.SuggestedFixes = append(result.SuggestedFixes, analysis.SuggestedFix{
			TextEdits: []analysis.TextEdit{
				{
					NewText: []byte(a.generateFix(style, vars)),
				},
			},
		})
		return result
	}

	return result
}

func (a *Analyzer) generateFix(style CommentStyleType, vals map[string]Value) string {
	// TODO: add values for quick fixes in config
	vals["YEAR_RANGE"] = vals["YEAR"]
	vals["MOD_YEAR_RANGE"] = vals["YEAR"]

	for _, v := range vals {
		_ = v.Calculate(vals)
	}

	fixTemplate, err := template.New("fix").Parse(a.template)
	if err != nil {
		return ""
	}
	fixOut := new(bytes.Buffer)
	_ = fixTemplate.Execute(fixOut, vals)
	res := fixOut.String()
	resSplit := strings.Split(res, "\n")

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

	return strings.Join(resSplit, "\n") + "\n"
}

func (a *Analyzer) getPerTargetValues(path string, file *ast.File) (map[string]Value, error) {
	var res = make(map[string]Value)

	for k, v := range a.values {
		res[k] = v
	}

	res["MOD_YEAR"] = a.values["YEAR"]
	res["MOD_YEAR_RANGE"] = a.values["YEAR_RANGE"]
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
	i := 0
	n := len(text)
	for i < n {
		// Check for template placeholder start
		if i+3 < n && text[i] == a.delimsLeft[0] && text[i+1] == a.delimsLeft[1] {
			// Find the end of the placeholder
			end := i + 2
			for end < n && !(text[end] == a.delimsRight[0] && end+1 < n && text[end+1] == a.delimsRight[1]) {
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
