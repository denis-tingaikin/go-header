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
	"go/ast"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer creates new analyzer based on template and goheader values
func NewAnalyzer(c *Config) *analysis.Analyzer {
	var initOncer sync.Once
	var initErr error
	var goheader *Analyzer
	return &analysis.Analyzer{
		Doc:  "the_only_doc",
		URL:  "https://github.com/denis-tingaikin/go-header",
		Name: "goheader",
		Run: func(p *analysis.Pass) (any, error) {
			initOncer.Do(func() {
				var templ string
				var vals map[string]Value

				templ, initErr = c.GetTemplate()
				if initErr != nil {
					return
				}

				vals, initErr = c.GetValues()
				if initErr != nil {
					return
				}
				goheader = New(WithTemplate(templ), WithValues(vals))
			})
			if initErr != nil {
				return nil, initErr
			}

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

						res := goheader.Analyze(&Target{
							File: file,
							Path: filename,
						})

						if res.Err == nil {
							continue
						}

						diag := analysis.Diagnostic{
							Pos:     0,
							Message: filename + ":" + res.Err.Error(),
						}

						if res.Fix != "" {

							diag.SuggestedFixes = []analysis.SuggestedFix{{
								TextEdits: []analysis.TextEdit{{
									Pos:     file.FileStart,
									End:     file.Package - 2,
									NewText: []byte(res.Fix),
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

// NewAnalyzerFromConfigPath creates a new analysis.Analyzer from goheader config file
func NewAnalyzerFromConfigPath(config *string) *analysis.Analyzer {
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
				goheader = NewAnalyzer(&cfg)
			})

			if err != nil {
				return nil, err
			}
			return goheader.Run(p)
		},
	}
}
