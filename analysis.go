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
	"go/token"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer creates a new analyzer based on template and goheader values
func NewAnalyzer(c *Config) (*analysis.Analyzer, error) {
	templ, err := c.GetTemplate()
	if err != nil {
		return nil, err
	}

	vals, err := c.GetValues()
	if err != nil {
		return nil, err
	}

	goheader := New(WithTemplate(templ), WithValues(vals), WithDelims(c.GetDelims()))

	return &analysis.Analyzer{
		Doc:              "Check file license header",
		URL:              "https://github.com/denis-tingaikin/go-header",
		Name:             "goheader",
		RunDespiteErrors: true,
		Run: func(pass *analysis.Pass) (any, error) {
			jobCh := make(chan *ast.File, len(pass.Files))

			for _, f := range pass.Files {
				file := f
				jobCh <- file
			}
			close(jobCh)

			var wg sync.WaitGroup

			for range c.GetParallel() {
				wg.Add(1)

				go func() {
					defer wg.Done()

					for file := range jobCh {
						filename := pass.Fset.PositionFor(file.Pos(), false).Filename
						if !strings.HasSuffix(filename, ".go") {
							continue
						}

						diag, err := goheader.Analyze(filename, file)
						if err != nil {
							// TODO handle the error.
							return
						}

						if diag == nil {
							continue
						}

						var line = 1
						if ast.IsGenerated(file) {
							line = 4
						}

						fileToken := pass.Fset.File(file.Pos())

						start := fileToken.LineStart(line)
						endLine := fileToken.Line(diag.End-diag.Pos+start) + 1

						var end token.Pos
						if endLine < fileToken.LineCount() {
							end = fileToken.LineStart(endLine)
						} else {
							end = start
						}

						diag.Pos = start
						diag.End = end

						if len(diag.SuggestedFixes) > 0 && len(diag.SuggestedFixes[0].TextEdits) > 0 {
							diag.SuggestedFixes[0].TextEdits[0].Pos = start
							diag.SuggestedFixes[0].TextEdits[0].End = end
						}

						pass.Report(*diag)
					}
				}()
			}

			wg.Wait()

			return nil, nil
		},
	}, nil
}
