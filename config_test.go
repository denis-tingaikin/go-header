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
	"os"
	"path/filepath"
	"testing"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetDelims(t *testing.T) {
	t.Run("default when empty", func(t *testing.T) {
		cfg := &goheader.Config{}
		assert.Equal(t, "{{}}", cfg.GetDelims())
	})

	t.Run("custom delimiters", func(t *testing.T) {
		cfg := &goheader.Config{Delims: "[[]]"}
		assert.Equal(t, "[[]]", cfg.GetDelims())
	})
}

func TestConfig_GetParallel(t *testing.T) {
	t.Run("default when zero", func(t *testing.T) {
		cfg := &goheader.Config{}
		assert.Greater(t, cfg.GetParallel(), 0)
	})

	t.Run("default when negative", func(t *testing.T) {
		cfg := &goheader.Config{Parallel: -1}
		assert.Greater(t, cfg.GetParallel(), 0)
	})

	t.Run("custom positive value", func(t *testing.T) {
		cfg := &goheader.Config{Parallel: 4}
		assert.Equal(t, 4, cfg.GetParallel())
	})
}

func TestConfig_GetTemplate(t *testing.T) {
	t.Run("inline template", func(t *testing.T) {
		cfg := &goheader.Config{Template: "hello {{ .YEAR }}"}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Contains(t, tmpl, "hello")
	})

	t.Run("empty template and path", func(t *testing.T) {
		cfg := &goheader.Config{}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Equal(t, "", tmpl)
	})

	t.Run("template from file path", func(t *testing.T) {
		dir := t.TempDir()
		tmplFile := filepath.Join(dir, "header.txt")
		err := os.WriteFile(tmplFile, []byte("Copyright {{ .YEAR }}"), 0o644)
		require.NoError(t, err)

		cfg := &goheader.Config{TemplatePath: tmplFile}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Contains(t, tmpl, "Copyright")
	})

	t.Run("template path file not found", func(t *testing.T) {
		cfg := &goheader.Config{TemplatePath: "/nonexistent/path/file.txt"}
		_, err := cfg.GetTemplate()
		require.Error(t, err)
	})
}

func TestConfig_GetValues(t *testing.T) {
	t.Run("empty config returns builtins", func(t *testing.T) {
		cfg := &goheader.Config{}
		vals, err := cfg.GetValues()
		require.NoError(t, err)
		assert.Contains(t, vals, "YEAR")
		assert.Contains(t, vals, "YEAR_RANGE")
	})

	t.Run("old style const values", func(t *testing.T) {
		cfg := &goheader.Config{
			Values: map[string]map[string]string{
				"const": {"company": "Acme"},
			},
		}
		vals, err := cfg.GetValues()
		require.NoError(t, err)
		assert.Contains(t, vals, "COMPANY")
		assert.Contains(t, vals, "company")
		assert.Equal(t, "Acme", vals["COMPANY"].Get())
	})

	t.Run("old style regexp values", func(t *testing.T) {
		cfg := &goheader.Config{
			Values: map[string]map[string]string{
				"regexp": {"holder": `[A-Z]+`},
			},
		}
		vals, err := cfg.GetValues()
		require.NoError(t, err)
		assert.Contains(t, vals, "HOLDER")
		assert.Contains(t, vals, "holder")
	})

	t.Run("vars style", func(t *testing.T) {
		cfg := &goheader.Config{
			Vars: map[string]string{"author": `Denis\s+T`},
		}
		vals, err := cfg.GetValues()
		require.NoError(t, err)
		assert.Contains(t, vals, "AUTHOR")
		assert.Contains(t, vals, "author")
	})
}

func TestConfig_FillSettings(t *testing.T) {
	t.Run("fills all fields", func(t *testing.T) {
		cfg := &goheader.Config{
			Template: "Copyright {{ .YEAR }}",
			Parallel: 2,
			Delims:   "[[]]",
			Experimental: goheader.Experimental{
				CGO: true,
			},
		}
		settings := &goheader.Settings{}
		err := cfg.FillSettings(settings)
		require.NoError(t, err)
		assert.Contains(t, settings.Template, "Copyright")
		assert.Equal(t, 2, settings.Parallel)
		assert.Equal(t, "[[", settings.LeftDelim)
		assert.Equal(t, "]]", settings.RightDelim)
		assert.True(t, settings.CGO)
	})

	t.Run("default delimiters when empty", func(t *testing.T) {
		cfg := &goheader.Config{Template: "test"}
		settings := &goheader.Settings{}
		err := cfg.FillSettings(settings)
		require.NoError(t, err)
		assert.Equal(t, "{{", settings.LeftDelim)
		assert.Equal(t, "}}", settings.RightDelim)
	})

	t.Run("odd length delimiters fall back to defaults", func(t *testing.T) {
		cfg := &goheader.Config{Template: "test", Delims: "abc"}
		settings := &goheader.Settings{}
		err := cfg.FillSettings(settings)
		require.NoError(t, err)
		assert.Equal(t, "{{", settings.LeftDelim)
		assert.Equal(t, "}}", settings.RightDelim)
	})

	t.Run("error on bad template path", func(t *testing.T) {
		cfg := &goheader.Config{TemplatePath: "/nonexistent"}
		settings := &goheader.Settings{}
		err := cfg.FillSettings(settings)
		require.Error(t, err)
	})
}

func TestParse(t *testing.T) {
	t.Run("valid yaml", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "config.yml")
		err := os.WriteFile(cfgFile, []byte("template: \"hello\"\nparallel: 3\n"), 0o644)
		require.NoError(t, err)

		cfg, err := goheader.Parse(cfgFile)
		require.NoError(t, err)
		assert.Equal(t, "hello", cfg.Template)
		assert.Equal(t, 3, cfg.Parallel)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := goheader.Parse("/nonexistent/file.yml")
		require.Error(t, err)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		cfgFile := filepath.Join(dir, "bad.yml")
		err := os.WriteFile(cfgFile, []byte(":::invalid:::yaml"), 0o644)
		require.NoError(t, err)

		_, err = goheader.Parse(cfgFile)
		require.Error(t, err)
	})
}

func TestSettings_SetTemplate(t *testing.T) {
	t.Run("inline template", func(t *testing.T) {
		s := &goheader.Settings{}
		err := s.SetTemplate("Copyright 2025", "")
		require.NoError(t, err)
		assert.Equal(t, "Copyright 2025", s.Template)
	})

	t.Run("empty both clears template", func(t *testing.T) {
		s := &goheader.Settings{Template: "old"}
		err := s.SetTemplate("", "")
		require.NoError(t, err)
		assert.Equal(t, "", s.Template)
	})

	t.Run("from file path", func(t *testing.T) {
		dir := t.TempDir()
		tmplFile := filepath.Join(dir, "tmpl.txt")
		err := os.WriteFile(tmplFile, []byte("  File Template  "), 0o644)
		require.NoError(t, err)

		s := &goheader.Settings{}
		err = s.SetTemplate("", tmplFile)
		require.NoError(t, err)
		assert.Equal(t, "File Template", s.Template)
	})

	t.Run("file not found", func(t *testing.T) {
		s := &goheader.Settings{}
		err := s.SetTemplate("", "/nonexistent/path")
		require.Error(t, err)
	})

	t.Run("inline takes precedence over path", func(t *testing.T) {
		s := &goheader.Settings{}
		err := s.SetTemplate("inline", "/some/path")
		require.NoError(t, err)
		assert.Equal(t, "inline", s.Template)
	})
}

func TestSettings_SetDelimiters(t *testing.T) {
	t.Run("custom delimiters", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetDelimiters("[[", "]]")
		assert.Equal(t, "[[", s.LeftDelim)
		assert.Equal(t, "]]", s.RightDelim)
	})

	t.Run("empty left defaults", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetDelimiters("", ">>")
		assert.Equal(t, "{{", s.LeftDelim)
		assert.Equal(t, ">>", s.RightDelim)
	})

	t.Run("empty right defaults", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetDelimiters("<<", "")
		assert.Equal(t, "<<", s.LeftDelim)
		assert.Equal(t, "}}", s.RightDelim)
	})

	t.Run("both empty defaults", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetDelimiters("", "")
		assert.Equal(t, "{{", s.LeftDelim)
		assert.Equal(t, "}}", s.RightDelim)
	})
}

func TestSettings_SetValues(t *testing.T) {
	t.Run("sets values with case variants", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetValues(map[string]string{
			"Company": `Acme\s+Corp`,
		})
		assert.Contains(t, s.Values, "COMPANY")
		assert.Contains(t, s.Values, "company")
		assert.Contains(t, s.Values, "YEAR")
		assert.Contains(t, s.Values, "YEAR_RANGE")
	})

	t.Run("empty map still has builtins", func(t *testing.T) {
		s := &goheader.Settings{}
		s.SetValues(map[string]string{})
		assert.Contains(t, s.Values, "YEAR")
		assert.Contains(t, s.Values, "YEAR_RANGE")
	})
}

func TestMigrateOldConfig(t *testing.T) {
	t.Run("migrates old style without dot", func(t *testing.T) {
		cfg := &goheader.Config{Template: "Copyright {{YEAR}}"}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Contains(t, tmpl, "{{ .YEAR }}")
	})

	t.Run("preserves new style with dot", func(t *testing.T) {
		cfg := &goheader.Config{Template: "Copyright {{ .YEAR }}"}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Contains(t, tmpl, "{{ .YEAR }}")
	})

	t.Run("migrates space-separated keys", func(t *testing.T) {
		cfg := &goheader.Config{Template: "{{ MY COMPANY }}"}
		tmpl, err := cfg.GetTemplate()
		require.NoError(t, err)
		assert.Contains(t, tmpl, "{{ .MY_COMPANY }}")
	})
}
