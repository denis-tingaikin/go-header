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

package goheader_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func newAnalyzer(t *testing.T, tmpl string) *goheader.Analyzer {
	t.Helper()
	cfg := goheader.Config{Template: tmpl}
	vals, err := cfg.GetValues()
	require.NoError(t, err)
	return &goheader.Analyzer{Settings: &goheader.Settings{
		Template:   tmpl,
		Values:     vals,
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}}
}

func header(t *testing.T, h string) (string, *ast.File) {
	t.Helper()
	return t.TempDir(), &ast.File{
		Comments: []*ast.CommentGroup{{
			List: []*ast.Comment{{Text: h}},
		}},
		Package: token.Pos(len(h)),
	}
}

func noHeader(t *testing.T) (string, *ast.File) {
	t.Helper()
	return t.TempDir(), &ast.File{Comments: nil, Package: 1}
}

func extractGolden(t *testing.T, filename string) string {
	t.Helper()

	fs := token.NewFileSet()
	tokenFile, err := parser.ParseFile(fs, filename, nil, parser.ParseComments)
	require.NoError(t, err)

	var h string
	for _, comment := range tokenFile.Comments[0].List {
		h += comment.Text + "\n"
	}

	return h
}

func TestAnalyzer(t *testing.T) {
	testCases := []struct {
		name        string
		cfgFilename string
	}{
		{name: "cgo", cfgFilename: "cgo.yml"},
		{name: "constvalue", cfgFilename: "constvalue.yml"},
		{name: "constvalue2", cfgFilename: "constvalue2.yml"},
		{name: "delimiters", cfgFilename: "delimiters.yml"},
		{name: "headercomment", cfgFilename: "headercomment.yml"},
		{name: "nestedvalues", cfgFilename: "nestedvalues.yml"},
		{name: "oldconfig", cfgFilename: "oldconfig.yml"},
		{name: "readme", cfgFilename: "readme.yml"},
		{name: "regexpvalue", cfgFilename: "regexpvalue.yml"},
		{name: "starcomment", cfgFilename: "starcomment.yml"},
		{name: "unicodeheader", cfgFilename: "unicodeheader.yml"},
		{name: "gobuild", cfgFilename: "gobuild.yml"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			testdata := analysistest.TestData()

			cfg, err := goheader.Parse(filepath.Join(testdata, "src", test.name, test.cfgFilename))
			require.NoError(t, err)

			cfg.Experimental.CGO = true

			settings := &goheader.Settings{}

			err = cfg.FillSettings(settings)
			require.NoError(t, err)

			analyzer := goheader.New(settings)

			analysistest.Run(t, testdata, analyzer, test.name)
		})
	}
}

func TestAnalyzer_fix(t *testing.T) {
	testCases := []struct {
		dir         string
		cfgFilename string
	}{
		{dir: "fix", cfgFilename: "fix.yml"},
		{dir: "sample", cfgFilename: "sample.yml"},
		{dir: "noheader", cfgFilename: "noheader.yml"},
		// {dir: "regexpvalue_issue", cfgFilename: "regexpvalue_issue.yml"}, // TODO: https://github.com/denis-tingaikin/go-header/issues/52
	}

	testdata := analysistest.TestData()

	for _, test := range testCases {
		list, err := os.ReadDir(filepath.Join(testdata, test.dir))
		require.NoError(t, err)

		for _, entry := range list {
			if !strings.HasSuffix(entry.Name(), ".go") || entry.IsDir() {
				continue
			}

			cfg, err := goheader.Parse(filepath.Join(testdata, test.dir, test.cfgFilename))
			require.NoError(t, err)

			cfg.Experimental.CGO = true

			settings := &goheader.Settings{}

			err = cfg.FillSettings(settings)
			require.NoError(t, err)

			t.Run(filepath.Join(test.dir, entry.Name()), func(t *testing.T) {
				gh := goheader.Analyzer{Settings: settings}

				srcFile := filepath.Join(testdata, test.dir, entry.Name())

				fs := token.NewFileSet()
				tokenFile, err := parser.ParseFile(fs, srcFile, nil, parser.ParseComments)
				require.NoError(t, err)

				diag, err := gh.Analyze(srcFile, tokenFile)
				require.NoError(t, err)

				require.NotNil(t, diag)
				require.Len(t, diag.SuggestedFixes, 1)
				require.Len(t, diag.SuggestedFixes[0].TextEdits, 1)

				expected := extractGolden(t, filepath.Join(testdata, test.dir, entry.Name()+".golden"))
				assert.Equal(t, expected, string(diag.SuggestedFixes[0].TextEdits[0].NewText))
			})
		}
	}
}

func TestAnalyzer_YearRangeValue_ShouldWorkWithComplexVariables(t *testing.T) {
	var cfg goheader.Config

	vals, err := cfg.GetValues()
	require.NoError(t, err)

	vals["MY_VAL"] = &goheader.RegexpValue{
		RawValue: "{{ .YEAR_RANGE }} B",
	}

	settings := &goheader.Settings{
		Values:     vals,
		Template:   "A {{ .MY_VAL }}",
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}

	a := goheader.Analyzer{Settings: settings}

	diag, err := a.Analyze(header(t, fmt.Sprintf("A 2000-%v B", time.Now().Year())))
	require.NoError(t, err)

	require.Nil(t, diag)
}

func TestNew_PanicsOnNilSettings(t *testing.T) {
	assert.Panics(t, func() {
		goheader.New(nil)
	})
}

func TestNew_ReturnsAnalyzer(t *testing.T) {
	a := goheader.New(&goheader.Settings{Parallel: 1})
	assert.Equal(t, "goheader", a.Name)
	assert.True(t, a.RunDespiteErrors)
}

func TestAnalyze_Matching(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	tests := []struct {
		name    string
		tmpl    string
		comment string
	}{
		{
			name:    "line comment matches",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "// Copyright " + year,
		},
		{
			name:    "block comment matches",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "/* Copyright " + year + " */",
		},
		{
			name:    "star block comment matches",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "/*\n * Copyright " + year + "\n */",
		},
		{
			name:    "template with regex metacharacters",
			tmpl:    "Copyright (c) {{ .YEAR }}",
			comment: "// Copyright (c) " + year,
		},
		{
			name:    "template with brackets and dots",
			tmpl:    "License [MIT] v1.0",
			comment: "// License [MIT] v1.0",
		},
		{
			name:    "template with plus and caret",
			tmpl:    "C++ header ^1.0",
			comment: "// C++ header ^1.0",
		},
		{
			name:    "template with pipe and dollar",
			tmpl:    "A|B end$",
			comment: "// A|B end$",
		},
		{
			name:    "template with backslash",
			tmpl:    `path\to\file`,
			comment: `// path\to\file`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAnalyzer(t, tt.tmpl)
			diag, err := a.Analyze(header(t, tt.comment))
			require.NoError(t, err)
			assert.Nil(t, diag)
		})
	}
}

func TestAnalyze_Mismatch(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    string
		comment string
	}{
		{
			name:    "wrong year",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "// Copyright WRONG",
		},
		{
			name:    "wrong text in block comment",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "/* Copyright WRONG */",
		},
		{
			name:    "wrong text in star block comment",
			tmpl:    "Copyright {{ .YEAR }}",
			comment: "/*\n * Copyright WRONG\n */",
		},
		{
			name:    "partial match",
			tmpl:    "Copyright {{ .YEAR }} All rights reserved",
			comment: fmt.Sprintf("// Copyright %d", time.Now().Year()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAnalyzer(t, tt.tmpl)
			diag, err := a.Analyze(header(t, tt.comment))
			require.NoError(t, err)
			require.NotNil(t, diag)
			assert.Equal(t, "template doesn't match", diag.Message)
		})
	}
}

func TestAnalyze_EmptyTemplate(t *testing.T) {
	a := &goheader.Analyzer{Settings: &goheader.Settings{
		Template:   "",
		Values:     map[string]goheader.Value{},
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}}

	diag, err := a.Analyze(header(t, "// some header"))
	require.NoError(t, err)
	assert.Nil(t, diag)
}

func TestAnalyze_NoComments_MissedHeader(t *testing.T) {
	a := newAnalyzer(t, "Copyright")
	diag, err := a.Analyze(noHeader(t))
	require.NoError(t, err)
	require.NotNil(t, diag)
	assert.Equal(t, "missed copyright header", diag.Message)
}

func TestAnalyze_MultilineTemplate(t *testing.T) {
	tmpl := "Copyright {{ .YEAR }}\nAll rights reserved"
	a := newAnalyzer(t, tmpl)

	path := t.TempDir()
	file := &ast.File{
		Comments: []*ast.CommentGroup{{
			List: []*ast.Comment{
				{Text: fmt.Sprintf("// Copyright %d", time.Now().Year())},
				{Text: "// All rights reserved"},
			},
		}},
		Package: 200,
	}

	diag, err := a.Analyze(path, file)
	require.NoError(t, err)
	assert.Nil(t, diag)
}

func TestAnalyze_Fix_DoubleSlash(t *testing.T) {
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")
	diag, err := a.Analyze(noHeader(t))
	require.NoError(t, err)
	require.NotNil(t, diag)
	require.Len(t, diag.SuggestedFixes, 1)

	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	assert.True(t, strings.HasPrefix(fix, "// "))
	assert.True(t, strings.HasSuffix(fix, "\n"))
	assert.Contains(t, fix, fmt.Sprint(time.Now().Year()))
}

func TestAnalyze_Fix_MultiLine(t *testing.T) {
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")
	diag, err := a.Analyze(header(t, "/* Copyright WRONG */"))
	require.NoError(t, err)
	require.Len(t, diag.SuggestedFixes, 1)

	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	assert.True(t, strings.HasPrefix(fix, "/*\n"))
	assert.True(t, strings.HasSuffix(fix, "*/\n"))
	assert.NotContains(t, fix, " * Copyright")
}

func TestAnalyze_Fix_MultiLineStar(t *testing.T) {
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")
	diag, err := a.Analyze(header(t, "/*\n * Copyright WRONG\n */"))
	require.NoError(t, err)
	require.Len(t, diag.SuggestedFixes, 1)

	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	assert.True(t, strings.HasPrefix(fix, "/*\n"))
	assert.Contains(t, fix, " * Copyright")
	assert.True(t, strings.HasSuffix(fix, " */\n"))
}

func TestAnalyze_Fix_MultilineTemplate_DoubleSlash(t *testing.T) {
	a := newAnalyzer(t, "Line one\nLine two")
	diag, err := a.Analyze(noHeader(t))
	require.NoError(t, err)
	require.Len(t, diag.SuggestedFixes, 1)

	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	lines := strings.Split(strings.TrimSuffix(fix, "\n"), "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "// Line one", lines[0])
	assert.Equal(t, "// Line two", lines[1])
}

func TestAnalyze_Fix_RegexpValueBlocksFix(t *testing.T) {
	a := &goheader.Analyzer{Settings: &goheader.Settings{
		Template: "Copyright {{ .HOLDER }}",
		Values: map[string]goheader.Value{
			"YEAR":       &goheader.ConstValue{RawValue: "2025"},
			"YEAR_RANGE": &goheader.RegexpValue{RawValue: `((20\d\d\-{{.YEAR}})|({{.YEAR}}))`},
			"HOLDER":     &goheader.RegexpValue{RawValue: `[A-Z]+`},
		},
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}}

	diag, err := a.Analyze(header(t, "// Copyright 12345"))
	require.NoError(t, err)
	require.NotNil(t, diag)
	assert.Equal(t, "template doesn't match", diag.Message)
	assert.Empty(t, diag.SuggestedFixes)
}

func TestAnalyze_SkipDirectives(t *testing.T) {
	tests := []struct {
		name       string
		directives []string
		hdr        string
		wantMatch  bool
	}{
		{
			name:       "+build directive skipped",
			directives: []string{"// +build linux"},
			hdr:        "// Copyright",
			wantMatch:  true,
		},
		{
			name:       "go:build directive skipped",
			directives: []string{"//go:build linux"},
			hdr:        "// Copyright",
			wantMatch:  true,
		},
		{
			name:       "cgo directive skipped",
			directives: []string{"// Code generated by cmd/cgo; DO NOT EDIT."},
			hdr:        "// Copyright",
			wantMatch:  true,
		},
		{
			name:       "empty comment skipped",
			directives: []string{"//"},
			hdr:        "// Copyright",
			wantMatch:  true,
		},
		{
			name:       "multiple directives all skipped",
			directives: []string{"//go:build linux", "// +build linux"},
			hdr:        "// Copyright",
			wantMatch:  true,
		},
		{
			name:       "no header after directives",
			directives: []string{},
			hdr:        "",
			wantMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAnalyzer(t, "Copyright")
			path := t.TempDir()

			var groups []*ast.CommentGroup
			for _, d := range tt.directives {
				groups = append(groups, &ast.CommentGroup{
					List: []*ast.Comment{{Text: d}},
				})
			}
			if tt.hdr != "" {
				groups = append(groups, &ast.CommentGroup{
					List: []*ast.Comment{{Text: tt.hdr}},
				})
			}

			file := &ast.File{Comments: groups, Package: 200}

			diag, err := a.Analyze(path, file)
			require.NoError(t, err)
			if tt.wantMatch {
				assert.Nil(t, diag)
			} else {
				require.NotNil(t, diag)
				assert.Equal(t, "missed copyright header", diag.Message)
			}
		})
	}
}

func TestAnalyze_QuoteMeta(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	tests := []struct {
		name          string
		tmpl          string
		matchComment  string
		rejectComment string
	}{
		{
			name:          "parentheses",
			tmpl:          "Copyright (c) {{ .YEAR }}",
			matchComment:  "// Copyright (c) " + year,
			rejectComment: "// Copyright Xc) " + year,
		},
		{
			name:          "square brackets",
			tmpl:          "License [MIT]",
			matchComment:  "// License [MIT]",
			rejectComment: "// License MIT",
		},
		{
			name:          "curly braces outside placeholder",
			tmpl:          "func() { {{ .YEAR }} }",
			matchComment:  "// func() { " + year + " }",
			rejectComment: "// func()   " + year + "  ",
		},
		{
			name:          "dot in text",
			tmpl:          "v1.0 {{ .YEAR }}",
			matchComment:  "// v1.0 " + year,
			rejectComment: "// v1X0 " + year,
		},
		{
			name:          "plus sign",
			tmpl:          "C++ {{ .YEAR }}",
			matchComment:  "// C++ " + year,
			rejectComment: "// Cpp " + year,
		},
		{
			name:          "asterisk",
			tmpl:          "* All rights",
			matchComment:  "// * All rights",
			rejectComment: "//  All rights",
		},
		{
			name:          "question mark",
			tmpl:          "Really?",
			matchComment:  "// Really?",
			rejectComment: "// Reall",
		},
		{
			name:          "pipe",
			tmpl:          "A|B",
			matchComment:  "// A|B",
			rejectComment: "// A",
		},
		{
			name:          "caret and dollar",
			tmpl:          "^start end$",
			matchComment:  "// ^start end$",
			rejectComment: "// start end",
		},
		{
			name:          "backslash",
			tmpl:          `path\to`,
			matchComment:  `// path\to`,
			rejectComment: "// pathXto",
		},
		{
			name:          "multiple metacharacters together",
			tmpl:          "f(x) = [a+b]*c",
			matchComment:  "// f(x) = [a+b]*c",
			rejectComment: "// fXxX = Xa+bX*c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/matches", func(t *testing.T) {
			a := newAnalyzer(t, tt.tmpl)
			diag, err := a.Analyze(header(t, tt.matchComment))
			require.NoError(t, err)
			assert.Nil(t, diag, "should match: %s", tt.matchComment)
		})
		t.Run(tt.name+"/rejects", func(t *testing.T) {
			a := newAnalyzer(t, tt.tmpl)
			diag, err := a.Analyze(header(t, tt.rejectComment))
			require.NoError(t, err)
			assert.NotNil(t, diag, "should reject: %s", tt.rejectComment)
		})
	}
}

func TestAnalyze_QuoteMeta_UnclosedPlaceholder(t *testing.T) {
	a := &goheader.Analyzer{Settings: &goheader.Settings{
		Template: "Copyright {{",
		Values: map[string]goheader.Value{
			"YEAR":       &goheader.ConstValue{RawValue: fmt.Sprint(time.Now().Year())},
			"YEAR_RANGE": &goheader.RegexpValue{RawValue: `((20\d\d\-{{.YEAR}})|({{.YEAR}}))`},
		},
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}}

	// Should not panic
	_, _ = a.Analyze(header(t, "// Copyright {{"))
}

func TestAnalyze_QuoteMeta_PlaceholderAtEnd(t *testing.T) {
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")
	year := fmt.Sprint(time.Now().Year())
	diag, err := a.Analyze(header(t, "// Copyright "+year))
	require.NoError(t, err)
	assert.Nil(t, diag)
}

func TestAnalyze_QuoteMeta_MultiplePlaceholders(t *testing.T) {
	a := newAnalyzer(t, "{{ .YEAR }} by {{ .YEAR }}")
	year := fmt.Sprint(time.Now().Year())
	diag, err := a.Analyze(header(t, "// "+year+" by "+year))
	require.NoError(t, err)
	assert.Nil(t, diag)
}

func TestAnalyze_StarBlock_Variants(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	tests := []struct {
		name      string
		comment   string
		wantMatch bool
	}{
		{
			name:      "standard star prefix",
			comment:   "/*\n * Copyright " + year + "\n */",
			wantMatch: true,
		},
		{
			name:      "bare asterisk no space",
			comment:   "/*\n *Copyright " + year + "\n */",
			wantMatch: true,
		},
		{
			name:      "plain block comment",
			comment:   "/* Copyright " + year + " */",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newAnalyzer(t, "Copyright {{ .YEAR }}")
			diag, err := a.Analyze(header(t, tt.comment))
			require.NoError(t, err)
			if tt.wantMatch {
				assert.Nil(t, diag)
			} else {
				assert.NotNil(t, diag)
			}
		})
	}
}

func TestAnalyze_ExtraContentAfterTemplate(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")

	// Unanchored match accepts extra content after the template
	diag, err := a.Analyze(header(t, "// Copyright "+year+" UNAUTHORIZED EXTRA"))
	require.NoError(t, err)
	_ = diag
}

func TestAnalyze_TemplateMatchesSubstring(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	a := newAnalyzer(t, "Copyright {{ .YEAR }}")

	// Unanchored match accepts extra content before and after
	diag, err := a.Analyze(header(t, "// PREFIX Copyright "+year+" SUFFIX"))
	require.NoError(t, err)
	_ = diag
}

func TestAnalyze_RepeatedCallsDoNotCorrupt(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	a := newAnalyzer(t, "Copyright {{ .YEAR_RANGE }}")

	// First call: missing header triggers generateFix
	diag1, err := a.Analyze(noHeader(t))
	require.NoError(t, err)
	require.NotNil(t, diag1)

	// Second call: year range should still work for matching
	diag2, err := a.Analyze(header(t, "// Copyright 2020-"+year))
	require.NoError(t, err)
	assert.Nil(t, diag2, "YEAR_RANGE should still match after a prior generateFix call")

	// Third call: another missing header
	diag3, err := a.Analyze(noHeader(t))
	require.NoError(t, err)
	require.NotNil(t, diag3)
	require.Len(t, diag3.SuggestedFixes, 1)
	fix := string(diag3.SuggestedFixes[0].TextEdits[0].NewText)
	assert.Contains(t, fix, year, "fix should still contain the current year")
}

func TestAnalyze_TemplateEndsWithOpenBraces(t *testing.T) {
	a := &goheader.Analyzer{Settings: &goheader.Settings{
		Template: "end{{",
		Values: map[string]goheader.Value{
			"YEAR":       &goheader.ConstValue{RawValue: fmt.Sprint(time.Now().Year())},
			"YEAR_RANGE": &goheader.RegexpValue{RawValue: `((20\d\d\-{{.YEAR}})|({{.YEAR}}))`},
		},
		LeftDelim:  "{{",
		RightDelim: "}}",
		Parallel:   1,
	}}

	assert.NotPanics(t, func() {
		_, _ = a.Analyze(header(t, "// end{{"))
	})
}

func TestAnalyze_PlaceholderIsEntireTemplate(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	a := newAnalyzer(t, "{{ .YEAR }}")

	diag, err := a.Analyze(header(t, "// "+year))
	require.NoError(t, err)
	assert.Nil(t, diag, "placeholder-only template should match")
}

func TestAnalyze_ConstValueNotTreatedAsRegex(t *testing.T) {
	cfg := &goheader.Config{
		Template: "Copyright {{ .COMPANY }}",
		Values: map[string]map[string]string{
			"const": {"COMPANY": "My.Corp"},
		},
	}
	settings := &goheader.Settings{}
	err := cfg.FillSettings(settings)
	require.NoError(t, err)

	a := &goheader.Analyzer{Settings: settings}

	// "My.Corp" — the dot should be literal, not regex wildcard
	diag, err := a.Analyze(header(t, "// Copyright My.Corp"))
	require.NoError(t, err)
	assert.Nil(t, diag, "exact match should pass")

	// "MyXCorp" should NOT match if the dot is treated literally
	diag, err = a.Analyze(header(t, "// Copyright MyXCorp"))
	require.NoError(t, err)
	assert.NotNil(t, diag, "MyXCorp should not match My.Corp — the dot should be literal")
}

func TestAnalyze_CustomDelimiters(t *testing.T) {
	cfg := goheader.Config{
		Template: "Copyright [[ .YEAR ]]",
		Delims:   "[[]]",
	}
	vals, err := cfg.GetValues()
	require.NoError(t, err)

	a := &goheader.Analyzer{Settings: &goheader.Settings{
		Template:   "Copyright [[ .YEAR ]]",
		Values:     vals,
		LeftDelim:  "[[",
		RightDelim: "]]",
		Parallel:   1,
	}}

	year := fmt.Sprint(time.Now().Year())
	diag, err := a.Analyze(header(t, "// Copyright "+year))
	require.NoError(t, err)
	assert.Nil(t, diag)
}
