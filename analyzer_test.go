// Copyright (c) 2020-2022 Denis Tingaikin
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
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"testing"

	goheader "github.com/denis-tingaikin/go-header"
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
func TestAnalyzer_YearRangeValue_ShouldWorkWithComplexVariables(t *testing.T) {
	var conf goheader.Configuration
	var vals, err = conf.GetValues()
	require.NoError(t, err)
	vals["my-val"] = &goheader.RegexpValue{
		RawValue: "{{year-range }} B",
	}
	var a = goheader.New(goheader.WithTemplate("A {{ my-val }}"), goheader.WithValues(vals))
	require.Nil(t, a.Analyze(header(`A 2000-2022 B`)))
}

func TestAnalyzer_Analyze1(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("A {{ YEAR }}\nB"),
		goheader.WithValues(map[string]goheader.Value{
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(`A 2020
B`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze2(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{COPYRIGHT HOLDER}}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT HOLDER": &goheader.RegexpValue{
				RawValue: "(A {{ YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(`A 2020
B
A 2020
B
TEXT
`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze3(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{COPYRIGHT HOLDER}}TEXT"),
		goheader.WithValues(map[string]goheader.Value{
			"COPYRIGHT HOLDER": &goheader.RegexpValue{
				RawValue: "(A {{ YEAR }}\n(.*)\n)+",
			},
			"YEAR": &goheader.ConstValue{
				RawValue: "2020",
			},
		}))
	issue := a.Analyze(header(`A 2020
B
A 2021
B
TEXT
`))
	require.NotNil(t, issue)
}

func TestAnalyzer_Analyze4(t *testing.T) {
	a := goheader.New(
		goheader.WithTemplate("{{ A }}"),
		goheader.WithValues(map[string]goheader.Value{
			"A": &goheader.RegexpValue{
				RawValue: "[{{ B }}{{ C }}]{{D}}",
			},
			"B": &goheader.ConstValue{
				RawValue: "a-",
			},
			"C": &goheader.RegexpValue{
				RawValue: "z",
			},
			"D": &goheader.ConstValue{
				RawValue: "{{E}}",
			},
			"E": &goheader.ConstValue{
				RawValue: "{7}",
			},
		}))
	issue := a.Analyze(header(`abcdefg`))
	require.Nil(t, issue)
}

func TestAnalyzer_Analyze5(t *testing.T) {
	a := goheader.New(goheader.WithTemplate("abc"))
	p := path.Join(os.TempDir(), t.Name()+".go")
	defer func() {
		_ = os.Remove(p)
	}()
	err := ioutil.WriteFile(p, []byte("/*abc*///comment\npackage abc"), os.ModePerm)
	require.Nil(t, err)
	s := token.NewFileSet()
	f, err := parser.ParseFile(s, p, nil, parser.ParseComments)
	require.Nil(t, err)
	require.Nil(t, a.Analyze(&goheader.Target{File: f, Path: p}))
}

func TestREADME(t *testing.T) {
	a := goheader.New(
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
				RawValue: "mycompany.com",
			},
		}))
	issue := a.Analyze(header(`mycompany.com
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
