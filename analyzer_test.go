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
	"path"
	"testing"
	"time"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_YearRangeValue_ShouldWorkWithComplexVariables(t *testing.T) {
	var conf goheader.Configuration
	var vals, err = conf.GetValues()
	require.NoError(t, err)

	vals["my-val"] = &goheader.RegexpValue{
		Key:      "my-val",
		RawValue: "{{ year-range }} B",
	}
	a, err := goheader.New(goheader.WithTemplate("A {{ my-val }}"), goheader.WithValues(vals))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(fmt.Sprintf(`A 2000-%d B`, time.Now().Year())))
	require.Nil(t, issue)
}

func TestAnalyzer_YearRangeValue_CurrentYearOnly(t *testing.T) {
	var conf goheader.Configuration
	var vals, err = conf.GetValues()
	require.NoError(t, err)

	vals["current-year"] = &goheader.RegexpValue{
		Key:      "current-year",
		RawValue: "{{ year-range }} C",
	}
	a, err := goheader.New(goheader.WithTemplate("A {{ current-year }}"), goheader.WithValues(vals))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(fmt.Sprintf(`A %d C`, time.Now().Year())))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze1(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate("A {{ YEAR }}\nB"),
		goheader.WithValues(map[string]goheader.Value{
			"YEAR": &goheader.ConstValue{
				Key:      "YEAR",
				RawValue: "2020",
			},
		}))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(`A 2020
B`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze2(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate("{{COPYRIGHT HOLDER}}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT HOLDER": &goheader.RegexpValue{
				Key:      "COPYRIGHT HOLDER",
				RawValue: "(A {{ YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				Key:      "YEAR",
				RawValue: "2020",
			},
		}))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(`A 2020
B
A 2020
B
TEXT
`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze3(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate("{{COPYRIGHT HOLDER}}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT HOLDER": &goheader.RegexpValue{
				Key:      "COPYRIGHT HOLDER",
				RawValue: "(A {{ YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				Key:      "YEAR",
				RawValue: "2020",
			},
		}))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(`A 2020
B
A 2021
B
TEXT
`))
	require.NotNil(t, issue)
}

func TestAnalyzer_Analyze4(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate("{{ A }}"),
		goheader.WithValues(map[string]goheader.Value{
			"A": &goheader.RegexpValue{
				Key:      "A",
				RawValue: "[{{ B }}{{ C }}]{{D}}",
			},
			"B": &goheader.ConstValue{
				Key:      "B",
				RawValue: "a-",
			},
			"C": &goheader.RegexpValue{
				Key:      "C",
				RawValue: "z",
			},
			"D": &goheader.ConstValue{
				Key:      "D",
				RawValue: "{{E}}",
			},
			"E": &goheader.ConstValue{
				Key:      "E",
				RawValue: "{7}",
			},
		}))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(`abcdefg`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze5(t *testing.T) {
	a, err := goheader.New(goheader.WithTemplate("abc"))
	require.NoError(t, err)

	p := path.Join(os.TempDir(), t.Name()+".go")
	t.Cleanup(func() {
		_ = os.Remove(p)
	})
	err = os.WriteFile(p, []byte("/*abc*///comment\npackage abc"), os.ModePerm)
	require.NoError(t, err)

	s := token.NewFileSet()
	f, err := parser.ParseFile(s, p, nil, parser.ParseComments)
	require.NoError(t, err)
	require.Nil(t, a.Analyze(&goheader.Target{File: f, Path: p}))
}

func TestAnalyzer_Analyze6(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate("A {{ some-value }} B"),
		goheader.WithValues(map[string]goheader.Value{
			"SOME-VALUE": &goheader.ConstValue{
				Key:      "SOME-VALUE",
				RawValue: "{{ some-value }}",
			},
		}),
	)
	require.ErrorIs(t, err, goheader.ErrRecursiveValue)
	require.Contains(t, err.Error(), "SOME-VALUE")
	require.Nil(t, a)
}

func TestREADME(t *testing.T) {
	a, err := goheader.New(
		goheader.WithTemplate(`{{ MY COMPANY }}
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
			"MY COMPANY": &goheader.ConstValue{
				Key:      "MY COMPANY",
				RawValue: "mycompany.com",
			},
		}))
	require.NoError(t, err)

	issue := a.Analyze(newTargetHeader(`mycompany.com
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
	require.Nil(t, issue)
}

func TestIssueFixing(t *testing.T) {
	const pkg = `

// Package foo
package foo

func Foo() { println("Foo") }
`

	cases := []struct {
		name        string
		header      string
		expectedFix *goheader.Fix
	}{
		{
			name: "line comment",
			header: `// mycompany.net
// SPDX-License-Identifier: Foo`,
			expectedFix: &goheader.Fix{
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
			name: "block comment 1",
			header: `/* mycompany.net
SPDX-License-Identifier: Foo */`,
			expectedFix: &goheader.Fix{
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
			name: "block comment 2",
			header: `/*
mycompany.net
SPDX-License-Identifier: Foo */`,
			expectedFix: &goheader.Fix{
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
			name: "block comment 3",
			header: `/* mycompany.net
SPDX-License-Identifier: Foo
*/`,
			expectedFix: &goheader.Fix{
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
			name: "block comment 4",
			header: `/*

mycompany.net
SPDX-License-Identifier: Foo

*/`,
			expectedFix: &goheader.Fix{
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

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			a, err := goheader.New(
				goheader.WithTemplate(`{{ MY COMPANY }}
SPDX-License-Identifier: Foo`),
				goheader.WithValues(map[string]goheader.Value{
					"MY COMPANY": &goheader.ConstValue{
						Key:      "MY COMPANY",
						RawValue: "mycompany.com",
					},
				}))
			require.NoError(t, err)

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "foo.go", tt.header+pkg, parser.ParseComments)
			require.NoError(t, err)

			issue := a.Analyze(&goheader.Target{
				File: file,
				Path: t.TempDir(),
			})
			require.NotNil(t, issue)
			require.NotNil(t, issue.Fix())
			require.Equal(t, tt.expectedFix, issue.Fix())
		})
	}
}

func newTargetHeader(header string) *goheader.Target {
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
