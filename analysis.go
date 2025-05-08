// Copyright (c) 2025 Denis Tingaikin
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
	"go/ast"
	"go/token"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer creates new analyzer based on template and goheader values
func NewAnalyzer(templ string, vars map[string]Value) *analysis.Analyzer {
	var goheader = New(WithTemplate(templ), WithValues(vars))
	return &analysis.Analyzer{
		Doc:  "the_only_doc",
		URL:  "https://github.com/denis-tingaikin/go-header",
		Name: "goheader",
		Run: func(p *analysis.Pass) (any, error) {
			var wg sync.WaitGroup

			var jobCh = make(chan *ast.File, len(p.Files))

			for _, file := range p.Files {
				jobCh <- file
			}
			close(jobCh)

			for range runtime.NumCPU() {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for file := range jobCh {

						filename := p.Fset.Position(file.Pos()).Filename
						if !strings.HasSuffix(filename, ".go") {
							continue
						}

						issue := goheader.Analyze(&Target{
							File: file,
							Path: filename,
						})

						if issue == nil {
							continue
						}
						f := p.Fset.File(file.Pos())

						commentLine := 1
						var offset int

						// Inspired by https://github.com/denis-tingaikin/go-header/blob/4c75a6a2332f025705325d6c71fff4616aedf48f/analyzer.go#L85-L92
						if len(file.Comments) > 0 && file.Comments[0].Pos() < file.Package {
							if !strings.HasPrefix(file.Comments[0].List[0].Text, "/*") {
								// When the comment is "//" there is a one character offset.
								offset = 1
							}
							commentLine = p.Fset.PositionFor(file.Comments[0].Pos(), true).Line
						}

						// Skip issues related to build directives.
						// https://github.com/denis-tingaikin/go-header/issues/18
						if issue.Location().Position-offset < 0 {
							continue
						}

						diag := analysis.Diagnostic{
							Pos:     f.LineStart(issue.Location().Line+1) + token.Pos(issue.Location().Position-offset), // The position of the first divergence.
							Message: issue.Message(),
						}

						if fix := issue.Fix(); fix != nil {
							current := len(fix.Actual)
							for _, s := range fix.Actual {
								current += len(s)
							}

							start := f.LineStart(commentLine)

							end := start + token.Pos(current)

							header := strings.Join(fix.Expected, "\n") + "\n"

							// Adds an extra line between the package and the header.
							if end == file.Package {
								header += "\n"
							}

							diag.SuggestedFixes = []analysis.SuggestedFix{{
								TextEdits: []analysis.TextEdit{{
									Pos:     start,
									End:     end,
									NewText: []byte(header),
								}},
							}}
						}

						p.Report(diag)
					}
				}()
			}

			wg.Wait()
			return nil, nil
		},
	}
}

// NewAnalyzerFromConfig creates a new analysis.Analyzer from goheader config file
func NewAnalyzerFromConfig(config *string) *analysis.Analyzer {
	var goheaderOncer sync.Once
	var goheader *analysis.Analyzer

	return &analysis.Analyzer{
		Doc:  "the_only_doc",
		URL:  "https://github.com/denis-tingaikin/go-header",
		Name: "goheader",
		Run: func(p *analysis.Pass) (any, error) {
			var err error
			goheaderOncer.Do(func() {
				var cfg Config
				if err = cfg.Parse(*config); err != nil {
					return
				}
				templ, err := cfg.GetTemplate()
				if err != nil {
					return
				}
				vars, err := cfg.GetValues()
				if err != nil {
					return
				}
				goheader = NewAnalyzer(templ, vars)
			})

			if err != nil {
				return nil, err
			}
			return goheader.Run(p)
		},
	}
}
