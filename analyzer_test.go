// Copyright (c) 2020-2024 Denis Tingaikin
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
	"testing"
	"time"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func header(header string) *goheader.Target {
	return &goheader.Target{
		File: &ast.File{
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
		},
		Path: os.TempDir(),
	}
}

func TestAnalyzer_Analyze(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		config   string
		assert   assert.ValueAssertionFunc
	}{
		{
			desc:     "const value",
			filename: "constvalue/constvalue.go",
			config:   "constvalue/constvalue.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "regexp value",
			filename: "regexpvalue/regexpvalue.go",
			config:   "regexpvalue/regexpvalue.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "regexp value with issue",
			filename: "regexpvalue_issue/regexpvalue_issue.go",
			config:   "regexpvalue_issue/regexpvalue_issue.yml",
			assert:   assert.NotNil,
		},
		{
			desc:     "nested values",
			filename: "nestedvalues/nestedvalues.go",
			config:   "nestedvalues/nestedvalues.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "header comment",
			filename: "headercomment/headercomment.go",
			config:   "headercomment/headercomment.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "readme",
			filename: "readme/readme.go",
			config:   "readme/readme.yml",
			assert:   assert.Nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cfg := &goheader.Configuration{}

			err := cfg.Parse(filepath.Join("testdata", test.config))
			require.NoError(t, err)

			values, err := cfg.GetValues()
			require.NoError(t, err)

			tmpl, err := cfg.GetTemplate()
			require.NoError(t, err)

			a := goheader.New(
				goheader.WithValues(values),
				goheader.WithTemplate(tmpl),
			)

			filename := filepath.Join("testdata", test.filename)

			file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ParseComments)
			require.NoError(t, err)

			issue := a.Analyze(&goheader.Target{Path: filename, File: file})

			test.assert(t, issue)
		})
	}
}

func TestAnalyzer_Analyze_fix(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		config   string
		expected goheader.Fix
	}{
		{
			desc:     "Line comment",
			filename: "fix/linecomment.go",
			config:   "fix/fix.yml",
			expected: goheader.Fix{
				Actual: []string{
					"// mycompany.net",
					"// SPDX-License-Identifier: Foo",
				},
				Expected: []string{
					"// mycompany.com",
					"// SPDX-License-Identifier: Foo",
				},
			},
		},
		{
			desc:     "Block comment 1",
			filename: "fix/blockcomment1.go",
			config:   "fix/fix.yml",
			expected: goheader.Fix{
				Actual: []string{
					"/* mycompany.net",
					"SPDX-License-Identifier: Foo */",
				},
				Expected: []string{
					"/* mycompany.com",
					"SPDX-License-Identifier: Foo */",
				},
			},
		},
		{
			desc:     "Block comment 2",
			filename: "fix/blockcomment2.go",
			config:   "fix/fix.yml",
			expected: goheader.Fix{
				Actual: []string{
					"/*",
					"mycompany.net",
					"SPDX-License-Identifier: Foo */",
				},
				Expected: []string{
					"/*",
					"mycompany.com",
					"SPDX-License-Identifier: Foo */",
				},
			},
		},
		{
			desc:     "Block comment 3",
			filename: "fix/blockcomment3.go",
			config:   "fix/fix.yml",
			expected: goheader.Fix{
				Actual: []string{
					"/* mycompany.net",
					"SPDX-License-Identifier: Foo",
					"*/",
				},
				Expected: []string{
					"/* mycompany.com",
					"SPDX-License-Identifier: Foo",
					"*/",
				},
			},
		},
		{
			desc:     "Block comment 4",
			filename: "fix/blockcomment4.go",
			config:   "fix/fix.yml",
			expected: goheader.Fix{
				Actual: []string{
					"/*",
					"",
					"mycompany.net",
					"SPDX-License-Identifier: Foo",
					"",
					"*/",
				},
				Expected: []string{
					"/*",
					"",
					"mycompany.com",
					"SPDX-License-Identifier: Foo",
					"",
					"*/",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cfg := &goheader.Configuration{}

			err := cfg.Parse(filepath.Join("testdata", test.config))
			require.NoError(t, err)

			values, err := cfg.GetValues()
			require.NoError(t, err)

			tmpl, err := cfg.GetTemplate()
			require.NoError(t, err)

			a := goheader.New(
				goheader.WithValues(values),
				goheader.WithTemplate(tmpl),
			)

			filename := filepath.Join("testdata", test.filename)

			file, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.ParseComments)
			require.NoError(t, err)

			issue := a.Analyze(&goheader.Target{Path: filename, File: file})

			assert.Equal(t, test.expected.Actual, issue.Fix().Actual)
			assert.Equal(t, test.expected.Expected, issue.Fix().Expected)
		})
	}
}

func TestAnalyzer_YearRangeValue_ShouldWorkWithComplexVariables(t *testing.T) {
	var conf goheader.Configuration
	var vals, err = conf.GetValues()
	require.NoError(t, err)

	vals["my-val"] = &goheader.RegexpValue{
		RawValue: "{{year-range }} B",
	}

	var a = goheader.New(goheader.WithTemplate("A {{ my-val }}"), goheader.WithValues(vals))
	require.Nil(t, a.Analyze(header(fmt.Sprintf("A 2000-%v B", time.Now().Year()))))
}
