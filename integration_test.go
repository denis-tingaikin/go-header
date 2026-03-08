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
	"golang.org/x/tools/go/analysis"
)

func analyzerFromConfig(t *testing.T, cfg *goheader.Config) *goheader.Analyzer {
	t.Helper()
	settings := &goheader.Settings{}
	err := cfg.FillSettings(settings)
	require.NoError(t, err)
	return &goheader.Analyzer{Settings: settings}
}

func analyzeSource(t *testing.T, a *goheader.Analyzer, src string) *analysis.Diagnostic {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	err := os.WriteFile(path, []byte(src), 0o644)
	require.NoError(t, err)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	require.NoError(t, err)

	diag, err := a.Analyze(path, file)
	require.NoError(t, err)
	return diag
}

func TestFixRoundtrip(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	tests := []struct {
		name   string
		cfg    *goheader.Config
		source string
	}{
		{
			name:   "missing header gets line comment fix",
			cfg:    &goheader.Config{Template: "Copyright " + year},
			source: "package foo\n",
		},
		{
			name: "missing header with const values",
			cfg: &goheader.Config{
				Template: "Copyright {{ .YEAR }} MyCompany",
				Values: map[string]map[string]string{
					"const": {"YEAR": year},
				},
			},
			source: "package foo\n",
		},
		{
			name:   "wrong line comment header gets replaced",
			cfg:    &goheader.Config{Template: "Copyright " + year},
			source: "// Copyright WRONG\n\npackage foo\n",
		},
		{
			name:   "wrong block comment header gets replaced",
			cfg:    &goheader.Config{Template: "Copyright " + year},
			source: "/* Copyright WRONG */\n\npackage foo\n",
		},
		{
			name:   "wrong star block comment header gets replaced",
			cfg:    &goheader.Config{Template: "Copyright " + year},
			source: "/*\n * Copyright WRONG\n */\n\npackage foo\n",
		},
		{
			name: "multiline template roundtrips",
			cfg: &goheader.Config{
				Template: "Copyright " + year + "\nSPDX-License-Identifier: Apache-2.0",
			},
			source: "package foo\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := analyzerFromConfig(t, tt.cfg)

			diag := analyzeSource(t, a, tt.source)
			require.NotNil(t, diag, "expected a diagnostic")
			require.Len(t, diag.SuggestedFixes, 1)
			require.Len(t, diag.SuggestedFixes[0].TextEdits, 1)

			fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
			require.NotEmpty(t, fix)

			// Re-analyze with the fix applied as the header.
			newSource := fix + "\npackage foo\n"
			diag2 := analyzeSource(t, a, newSource)
			assert.Nil(t, diag2, "fix should satisfy the template, got: %v", diag2)
		})
	}
}

func TestConfigMigration_CompanyRename(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	oldCfg := &goheader.Config{
		Template: "Copyright " + year + " OldCorp",
	}
	newCfg := &goheader.Config{
		Template: "Copyright " + year + " NewCorp",
	}

	src := "// Copyright " + year + " OldCorp\n\npackage foo\n"

	// Passes old config.
	oldA := analyzerFromConfig(t, oldCfg)
	diag := analyzeSource(t, oldA, src)
	assert.Nil(t, diag, "should pass old config")

	// Fails new config.
	newA := analyzerFromConfig(t, newCfg)
	diag = analyzeSource(t, newA, src)
	require.NotNil(t, diag, "should fail new config")
	assert.Equal(t, "template doesn't match", diag.Message)

	// Fix contains the new company name.
	require.Len(t, diag.SuggestedFixes, 1)
	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	assert.Contains(t, fix, "NewCorp")
	assert.NotContains(t, fix, "OldCorp")
}

func TestConfigMigration_AddSPDX(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	simpleCfg := &goheader.Config{
		Template: "Copyright " + year + " MyCorp",
	}
	spdxCfg := &goheader.Config{
		Template: "Copyright " + year + " MyCorp\nSPDX-License-Identifier: Apache-2.0",
	}

	src := "// Copyright " + year + " MyCorp\n\npackage foo\n"

	// Passes the simple config.
	diag := analyzeSource(t, analyzerFromConfig(t, simpleCfg), src)
	assert.Nil(t, diag)

	// Fails the SPDX config.
	diag = analyzeSource(t, analyzerFromConfig(t, spdxCfg), src)
	require.NotNil(t, diag)
	assert.Equal(t, "template doesn't match", diag.Message)

	// Fix includes both lines.
	require.Len(t, diag.SuggestedFixes, 1)
	fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
	assert.Contains(t, fix, "Copyright")
	assert.Contains(t, fix, "SPDX-License-Identifier: Apache-2.0")
}

func TestConfigMigration_OldSyntaxToNewSyntax(t *testing.T) {
	oldCfg := &goheader.Config{
		Template: "{{ MY COMPANY }}",
		Values: map[string]map[string]string{
			"const": {"MY_COMPANY": "Acme Inc"},
		},
	}

	newCfg := &goheader.Config{
		Template: "{{ .MY_COMPANY }}",
		Vars:     map[string]string{"MY_COMPANY": "Acme Inc"},
	}

	src := "// Acme Inc\n\npackage foo\n"

	for name, cfg := range map[string]*goheader.Config{"old": oldCfg, "new": newCfg} {
		t.Run(name+"_syntax_accepts", func(t *testing.T) {
			diag := analyzeSource(t, analyzerFromConfig(t, cfg), src)
			assert.Nil(t, diag)
		})
	}
}

func TestConfigMigration_CustomDelimitersToDefault(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	customCfg := &goheader.Config{
		Template: "Copyright [[ .YEAR ]]",
		Delims:   "[[]]",
		Vars:     map[string]string{"YEAR": year},
	}

	defaultCfg := &goheader.Config{
		Template: "Copyright {{ .YEAR }}",
		Vars:     map[string]string{"YEAR": year},
	}

	src := "// Copyright " + year + "\n\npackage foo\n"

	for name, cfg := range map[string]*goheader.Config{"custom_delims": customCfg, "default_delims": defaultCfg} {
		t.Run(name, func(t *testing.T) {
			diag := analyzeSource(t, analyzerFromConfig(t, cfg), src)
			assert.Nil(t, diag)
		})
	}
}

func TestHeaderStylePreservation(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	tmpl := "Copyright " + year

	tests := []struct {
		name      string
		source    string
		fixPrefix string
		fixSuffix string
	}{
		{
			name:      "line comment style preserved",
			source:    "// Copyright WRONG\n\npackage foo\n",
			fixPrefix: "// ",
			fixSuffix: "\n",
		},
		{
			name:      "block comment style preserved",
			source:    "/* Copyright WRONG */\n\npackage foo\n",
			fixPrefix: "/*\n",
			fixSuffix: "*/\n",
		},
		{
			name:      "star block comment style preserved",
			source:    "/*\n * Copyright WRONG\n */\n\npackage foo\n",
			fixPrefix: "/*\n",
			fixSuffix: " */\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &goheader.Config{Template: tmpl}
			diag := analyzeSource(t, analyzerFromConfig(t, cfg), tt.source)
			require.NotNil(t, diag)
			require.Len(t, diag.SuggestedFixes, 1)

			fix := string(diag.SuggestedFixes[0].TextEdits[0].NewText)
			assert.True(t, strings.HasPrefix(fix, tt.fixPrefix),
				"fix should start with %q, got %q", tt.fixPrefix, fix)
			assert.True(t, strings.HasSuffix(fix, tt.fixSuffix),
				"fix should end with %q, got %q", tt.fixSuffix, fix)
		})
	}
}

func TestYearRange_AcceptsCurrentAndHistoricYears(t *testing.T) {
	cfg := &goheader.Config{
		Template: "Copyright {{ .YEAR_RANGE }} MyCorp",
	}
	a := analyzerFromConfig(t, cfg)
	year := fmt.Sprint(time.Now().Year())

	tests := []struct {
		name    string
		source  string
		wantNil bool
	}{
		{
			name:    "current year only",
			source:  "// Copyright " + year + " MyCorp\n\npackage foo\n",
			wantNil: true,
		},
		{
			name:    "year range ending in current year",
			source:  "// Copyright 2020-" + year + " MyCorp\n\npackage foo\n",
			wantNil: true,
		},
		{
			name:    "wrong current year",
			source:  "// Copyright 1999 MyCorp\n\npackage foo\n",
			wantNil: false,
		},
		{
			name:    "range ending in wrong year",
			source:  "// Copyright 2020-1999 MyCorp\n\npackage foo\n",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := analyzeSource(t, a, tt.source)
			if tt.wantNil {
				assert.Nil(t, diag)
			} else {
				assert.NotNil(t, diag)
			}
		})
	}
}

func TestFullLicenseHeader(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())

	cfg := &goheader.Config{
		Template: "Copyright (c) {{ .YEAR_RANGE }} {{ .COMPANY }}\n" +
			"\n" +
			"SPDX-License-Identifier: Apache-2.0\n" +
			"\n" +
			"Licensed under the Apache License, Version 2.0 (the \"License\");\n" +
			"you may not use this file except in compliance with the License.",
		Vars: map[string]string{
			"COMPANY": "MyCorp(, Inc\\.)?",
		},
	}
	a := analyzerFromConfig(t, cfg)

	tests := []struct {
		name    string
		company string
		wantNil bool
	}{
		{name: "with Inc", company: "MyCorp, Inc.", wantNil: true},
		{name: "without Inc", company: "MyCorp", wantNil: true},
		{name: "wrong company", company: "WrongCorp", wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := fmt.Sprintf(
				"// Copyright (c) %s %s\n"+
					"//\n"+
					"// SPDX-License-Identifier: Apache-2.0\n"+
					"//\n"+
					"// Licensed under the Apache License, Version 2.0 (the \"License\");\n"+
					"// you may not use this file except in compliance with the License.\n\n"+
					"package foo\n",
				year, tt.company,
			)

			diag := analyzeSource(t, a, src)
			if tt.wantNil {
				assert.Nil(t, diag)
			} else {
				assert.NotNil(t, diag)
			}
		})
	}
}

func TestBuildDirectivesWithHeaderEnforcement(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	cfg := &goheader.Config{Template: "Copyright " + year}
	a := analyzerFromConfig(t, cfg)

	tests := []struct {
		name    string
		source  string
		wantNil bool
	}{
		{
			name:    "build directive with correct header",
			source:  "//go:build linux\n\n// Copyright " + year + "\n\npackage foo\n",
			wantNil: true,
		},
		{
			name:    "build directive with wrong header",
			source:  "//go:build linux\n\n// Copyright WRONG\n\npackage foo\n",
			wantNil: false,
		},
		{
			name:    "build directive with no header",
			source:  "//go:build linux\n\npackage foo\n",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := analyzeSource(t, a, tt.source)
			if tt.wantNil {
				assert.Nil(t, diag)
			} else {
				assert.NotNil(t, diag)
			}
		})
	}
}

func TestConfigFromYAML_EndToEnd(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	dir := t.TempDir()

	cfgContent := fmt.Sprintf(`template: |-
  Copyright %s {{ .COMPANY }}
  SPDX-License-Identifier: MIT
vars:
  COMPANY: 'TestCorp'
`, year)

	cfgPath := filepath.Join(dir, "go-header.yml")
	err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	cfg, err := goheader.Parse(cfgPath)
	require.NoError(t, err)
	a := analyzerFromConfig(t, cfg)

	validSrc := fmt.Sprintf("// Copyright %s TestCorp\n// SPDX-License-Identifier: MIT\n\npackage foo\n", year)
	diag := analyzeSource(t, a, validSrc)
	assert.Nil(t, diag)

	invalidSrc := fmt.Sprintf("// Copyright %s WrongCorp\n// SPDX-License-Identifier: MIT\n\npackage foo\n", year)
	diag = analyzeSource(t, a, invalidSrc)
	assert.NotNil(t, diag)
}

func TestConfigFromYAML_TemplatePath(t *testing.T) {
	year := fmt.Sprint(time.Now().Year())
	dir := t.TempDir()

	tmplContent := "Copyright " + year + " FileCorp"
	tmplPath := filepath.Join(dir, "header.tmpl")
	err := os.WriteFile(tmplPath, []byte(tmplContent), 0o644)
	require.NoError(t, err)

	cfgContent := fmt.Sprintf("template-path: %s\n", tmplPath)
	cfgPath := filepath.Join(dir, "go-header.yml")
	err = os.WriteFile(cfgPath, []byte(cfgContent), 0o644)
	require.NoError(t, err)

	cfg, err := goheader.Parse(cfgPath)
	require.NoError(t, err)
	a := analyzerFromConfig(t, cfg)

	diag := analyzeSource(t, a, "// Copyright "+year+" FileCorp\n\npackage foo\n")
	assert.Nil(t, diag)

	diag = analyzeSource(t, a, "// Copyright "+year+" OtherCorp\n\npackage foo\n")
	assert.NotNil(t, diag)
}
