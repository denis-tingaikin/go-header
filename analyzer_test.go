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
	"path"
	"path/filepath"
	"testing"
	"time"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func header(t *testing.T, header string) *goheader.Target {
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
		Path: t.TempDir(),
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
			desc:     "const value 2",
			filename: "constvalue2/constvalue.go",
			config:   "constvalue2/constvalue.yml",
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
			desc:     "golangci-linter sample",
			filename: "golangci-linter/sample.go",
			config:   "golangci-linter/sample.yml",
			assert:   assert.NotNil,
		},
		{
			desc:     "nested values",
			filename: "nestedvalues/nestedvalues.go",
			config:   "nestedvalues/nestedvalues.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "missed header ",
			filename: "noheader/noheader.go",
			config:   "noheader/noheader.yml",
			assert:   assert.NotNil,
		},
		{
			desc:     "headercomment",
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
		{
			desc:     "cgo",
			filename: "cgo/cgo.go",
			config:   "cgo/cgo.yml",
			assert:   assert.NotNil,
		},
		{
			desc:     "star-block like header",
			filename: "starcomment/starcomment.go",
			config:   "starcomment/starcomment.yml",
			assert:   assert.Nil,
		},
		{
			desc:     "checks old config compatibility",
			filename: "oldconfig/oldconfig.go",
			config:   "oldconfig/oldconfig.yml",
			assert:   assert.Nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cfg := &goheader.Config{}

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

			test.assert(t, issue.Err)
		})
	}
}

func TestAnalyzer_Analyze_fix(t *testing.T) {
	testCases := []struct {
		desc     string
		filename string
		config   string
		expected goheader.Result
	}{
		{
			desc:     "Line comment",
			filename: "fix/linecomment.go",
			config:   "fix/fix.yml",
			expected: goheader.Result{
				Fix: `// mycompany.com
// SPDX-License-Identifier: Foo`,
			},
		},
		{
			desc:     "Block comment 1",
			filename: "fix/blockcomment1.go",
			config:   "fix/fix.yml",
			expected: goheader.Result{
				Fix: `/*
mycompany.com
SPDX-License-Identifier: Foo
*/`,
			},
		},
		{
			desc:     "Block comment 2",
			filename: "fix/blockcomment2.go",
			config:   "fix/fix.yml",

			expected: goheader.Result{
				Fix: `/*
mycompany.com
SPDX-License-Identifier: Foo
*/`,
			},
		},
		{
			desc:     "Block comment 3",
			filename: "fix/blockcomment3.go",
			config:   "fix/fix.yml",
			expected: goheader.Result{
				Fix: `/*
mycompany.com
SPDX-License-Identifier: Foo
*/`,
			},
		},
		{
			desc:     "Block comment 4",
			filename: "fix/blockcomment4.go",
			config:   "fix/fix.yml",
			expected: goheader.Result{
				Fix: `/*
mycompany.com
SPDX-License-Identifier: Foo
*/`,
			},
		},
		{
			desc:     "Star block comment",
			filename: "fix/blockcomment5.go",
			config:   "fix/fix.yml",
			expected: goheader.Result{
				Fix: `/*
 * mycompany.com
 * SPDX-License-Identifier: Foo
 */`,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			cfg := &goheader.Config{}

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

			actual := a.Analyze(&goheader.Target{Path: filename, File: file})

			assert.Equal(t, test.expected.Fix, actual.Fix)
		})
	}
}

func TestAnalyzer_YearRangeValue_ShouldWorkWithComplexVariables(t *testing.T) {
	var conf goheader.Config
	var vals, err = conf.GetValues()
	require.NoError(t, err)

	vals["MY_VAL"] = &goheader.RegexpValue{
		RawValue: "{{ .YEAR_RANGE }} B",
	}

	var a = goheader.New(goheader.WithTemplate("A {{ .MY_VAL }}"), goheader.WithValues(vals))
	require.Nil(t, a.Analyze(header(t, fmt.Sprintf("A 2000-%v B", time.Now().Year()))).Err)
}

func TestAnalyzer_UnicodeHeaders(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("😊早安😊"),
	)
	issue := a.Analyze(header(t, `😊早安😊`))
	require.Nil(t, issue.Err)
}

func TestAnalyzer_Analyze1(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("A {{ .YEAR }}\nB"),
		goheader.WithValues(map[string]goheader.Value{
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(t, `A 2020
B`))
	require.Nil(t, issue.Err)
}

func TestAnalyzer_Analyze2(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{ .COPYRIGHT_HOLDER }}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT_HOLDER": &goheader.RegexpValue{
				RawValue: "(A {{ .YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(t, `A 2020
B
A 2020
B
TEXT
`))
	require.Nil(t, issue.Err)
}

func TestAnalyzer_Analyze3(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{.COPYRIGHT_HOLDER}}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT_HOLDER": &goheader.RegexpValue{
				RawValue: "(A {{ .YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(t, `A 2020
B
A 2021
B
TEXT
`))
	require.NotNil(t, issue.Err)
}

func TestAnalyzer_Analyze4(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{ .A }}"),
		goheader.WithValues(map[string]goheader.Value{
			"A": &goheader.RegexpValue{
				RawValue: "[{{ .B }}{{ .C }}]{{.D}}",
			},
			"B": &goheader.ConstValue{
				RawValue: "a-",
			},
			"C": &goheader.RegexpValue{
				RawValue: "z",
			},
			"D": &goheader.ConstValue{
				RawValue: "{{.E}}",
			},
			"E": &goheader.ConstValue{
				RawValue: "{7}",
			},
		}))
	issue := a.Analyze(header(t, `abcdefg`))
	require.Nil(t, issue.Err)
}

func TestAnalyzer_Analyze5(t *testing.T) {
	a := goheader.New(goheader.WithTemplate("abc"))
	p := path.Join(os.TempDir(), t.Name()+".go")
	defer func() {
		_ = os.Remove(p)
	}()
	err := os.WriteFile(p, []byte("/*abc*/\n\n//comment\npackage abc"), os.ModePerm)
	require.Nil(t, err)
	s := token.NewFileSet()
	f, err := parser.ParseFile(s, p, nil, parser.ParseComments)
	require.Nil(t, err)
	require.Nil(t, a.Analyze(&goheader.Target{File: f, Path: p}).Err)
}

func TestAnalyzer_Analyze6(t *testing.T) {
	a := goheader.New(goheader.WithTemplate("abc"))
	p := path.Join(t.TempDir(), t.Name()+".go")
	defer func() {
		_ = os.Remove(p)
	}()

	err := os.WriteFile(p, []byte("//abc\n\n//comment\npackage abc"), os.ModePerm)
	require.Nil(t, err)
	s := token.NewFileSet()
	f, err := parser.ParseFile(s, p, nil, parser.ParseComments)
	require.Nil(t, err)
	require.Nil(t, a.Analyze(&goheader.Target{File: f, Path: p}).Err)
}

func TestREADME(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate(`{{ .MY_COMPANY }}
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`),
		goheader.WithValues(map[string]goheader.Value{
			"MY_COMPANY": &goheader.ConstValue{
				RawValue: "mycompany.com",
			},
		}))
	issue := a.Analyze(header(t, `mycompany.com
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.`))
	require.Nil(t, issue.Err)
}
