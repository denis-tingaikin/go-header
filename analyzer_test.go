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

func extractGolden(t *testing.T, filename string) string {
	t.Helper()

	fs := token.NewFileSet()
	tokenFile, err := parser.ParseFile(fs, filename, nil, parser.ParseComments)
	require.NoError(t, err)

	var header string
	for _, comment := range tokenFile.Comments[0].List {
		header += comment.Text + "\n"
	}

	return header
}

func header(t *testing.T, header string) (string, *ast.File) {
	return t.TempDir(), &ast.File{
		Comments: []*ast.CommentGroup{
			{
				List: []*ast.Comment{
					{
						Text: header,
					},
				},
			},
		},

		Package: token.Pos(len(header)),
	}
}
