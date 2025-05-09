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
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type CommentStyleType int

const (
	DoubleSlash CommentStyleType = iota
	MultiLine
	MultiLineStar
)

type Target struct {
	Path string
	File *ast.File
}

const iso = "2006-01-02 15:04:05 -0700"

func (t *Target) ModTime() (time.Time, error) {
	diff, err := exec.Command("git", "diff", t.Path).CombinedOutput()
	if err == nil && len(diff) == 0 {
		line, err := exec.Command("git", "log", "-1", "--pretty=format:%cd", "--date=iso", "--", t.Path).CombinedOutput()
		if err == nil {
			return time.Parse(iso, string(line))
		}
	}
	info, err := os.Stat(t.Path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

type Analyzer struct {
	values   map[string]Value
	template string
}

func New(opts ...Option) *Analyzer {
	var a Analyzer
	for _, opt := range opts {
		opt.apply(&a)
	}
	return &a
}

type Result struct {
	Err error
	Fix string
}

func (a *Analyzer) Analyze(t *Target) *Result {
	file := t.File
	header := ""

	var style CommentStyleType

	if len(file.Comments) > 0 && file.Comments[0].Pos() < file.Package {
		if strings.HasPrefix(file.Comments[0].List[0].Text, "/*") {
			header = (&ast.CommentGroup{List: []*ast.Comment{file.Comments[0].List[0]}}).Text()
			style = MultiLine

			if handledHeader, ok := handleStarBlock(header); ok {
				header = handledHeader
				style = MultiLineStar
			}

		} else {
			style = DoubleSlash
			header = file.Comments[0].Text()
		}
	}
	header = strings.TrimSpace(header)

	vars, err := a.getPerTargetValues(t)
	if err != nil {
		return &Result{Err: err}
	}

	templateRaw := quoteMeta(a.template)

	template, err := template.New("header").Parse(templateRaw)
	if err != nil {
		return &Result{Err: err}
	}

	res := new(bytes.Buffer)

	if err := template.Execute(res, vars); err != nil {
		return &Result{Err: err}
	}

	headerTemplate := res.String()

	r, err := regexp.Compile(headerTemplate)

	if err != nil {
		return &Result{Err: err}
	}

	if !r.MatchString(header) {
		// log.Println(header)
		// log.Println("template " + headerTemplate)
		return &Result{Err: errors.New("template doens't match"), Fix: a.generateFix(style, vars)}
	}

	return &Result{}
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
	if style == MultiLine {
		resSplit[0] = "/* " + resSplit[0]
	}

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
	case MultiLine:
		resSplit[len(resSplit)-1] = resSplit[len(resSplit)-1] + " */"
	case MultiLineStar:
		resSplit = append([]string{"/*"}, resSplit...)
		resSplit = append(resSplit, " */")
	}

	return strings.Join(resSplit, "\n")
}

func (a *Analyzer) getPerTargetValues(target *Target) (map[string]Value, error) {
	var res = make(map[string]Value)

	for k, v := range a.values {
		res[k] = v
	}

	res["MOD_YEAR"] = a.values["YEAR"]
	res["MOD_YEAR_RANGE"] = a.values["YEAR_RANGE"]
	if t, err := target.ModTime(); err == nil {
		res["MOD_YEAR"] = &ConstValue{RawValue: fmt.Sprint(t.Year())}
		res["MOD_YEAR_RANGE"] = &RegexpValue{RawValue: `((20\d\d\-{{MOD_YEAR}})|({{MOD_YEAR}}))`}
	}

	for _, v := range res {
		if err := v.Calculate(res); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// TODO: Fix vibe conding
func quoteMeta(text string) string {
	var result strings.Builder
	i := 0
	n := len(text)

	for i < n {
		// Check for template placeholder start
		if i+3 < n && text[i] == '{' && text[i+1] == '{' {
			// Find the end of the placeholder
			end := i + 2
			for end < n && !(text[end] == '}' && end+1 < n && text[end+1] == '}') {
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

func isNewLineRequired(group *ast.CommentGroup) bool {
	if len(group.List) < 2 {
		return false
	}
	end := group.List[0].End()
	pos := group.List[1].Pos()
	return end+1 >= pos && group.List[0].Text[len(group.List[0].Text)-1] != '\n'
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
