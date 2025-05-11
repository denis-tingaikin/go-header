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
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents go-header linter setup parameters
type Config struct {
	// Values is map of values. Supports two types 'const` and `regexp`. Values can be used recursively.
	// DEPRECATED: Use Vars instead.
	Values map[string]map[string]string `yaml:"values"`
	// Template is template for checking. Uses values.
	Template string `yaml:"template"`
	// TemplatePath path to the template file. Useful if need to load the template from a specific file.
	TemplatePath string `yaml:"template-path"`
	// Vars is map of values. Values can be used recursively.
	Vars map[string]string `yaml:"vars"`
	// ValuesMarker represents a string of marker for values. Default is "{{}}"
	ValuesMarker string `yaml:"values-marker"`
	// Parallel means a number of goroutines to proccess files.
	Parallel int `yaml:"parallel"`
}

func (c *Config) GetValuesMarker() string {
	if c.ValuesMarker == "" {
		return "{{}}"
	}
	return c.ValuesMarker
}

func (c *Config) GetParallel() int {
	if c.Parallel <= 0 {
		return runtime.NumCPU()
	}
	return c.Parallel
}

func (c *Config) builtInValues() map[string]Value {
	var result = make(map[string]Value)
	year := fmt.Sprint(time.Now().Year())
	result["YEAR_RANGE"] = &RegexpValue{
		RawValue: `((20\d\d\-{{.YEAR}})|({{.YEAR}}))`,
	}
	result["YEAR"] = &ConstValue{
		RawValue: year,
	}
	return result
}

func (c *Config) GetValues() (map[string]Value, error) {
	var result = c.builtInValues()
	createConst := func(raw string) Value {
		return &ConstValue{RawValue: raw}
	}
	createRegexp := func(raw string) Value {
		return &RegexpValue{RawValue: raw}
	}
	appendValues := func(m map[string]string, create func(string) Value) {
		for k, v := range m {
			result[strings.ToLower(k)] = create(v)
			result[strings.ToUpper(k)] = create(v)
		}
	}
	appendValues(c.Values["const"], createConst)
	appendValues(c.Values["regexp"], createRegexp)
	appendValues(c.Vars, createRegexp)
	return result, nil
}

func (c *Config) GetTemplate() (string, error) {
	var tmpl, err = c.getTemplate()

	if err != nil {
		return tmpl, err
	}

	return convertPlaceholders(tmpl, c.GetValuesMarker()), nil
}

func (c *Config) getTemplate() (string, error) {
	if c.Template != "" {
		return c.Template, nil
	}
	if c.TemplatePath == "" {
		return "", nil
	}
	if b, err := os.ReadFile(c.TemplatePath); err != nil {
		return "", err
	} else {
		c.Template = strings.TrimSpace(string(b))
		return c.Template, nil
	}
}

func (c *Config) Parse(p string) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, c)
}

func convertPlaceholders(input, marker string) string {
	var left = marker[:len(marker)/2]
	var right = marker[len(marker)/2:]

	// Regular expression to find all {{...}} patterns
	re := regexp.MustCompile(regexp.QuoteMeta(left) + `\s*([^` + right[:1] + `]+)\s*` + regexp.QuoteMeta(right))

	// Replace each match with the converted version
	result := re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the inner content (between {{ and }})
		inner := match[2 : len(match)-2]
		inner = strings.TrimSpace(inner)

		if strings.HasPrefix(inner, ".") {
			return "{{ " + inner + " }}"
		}

		// Replace spaces with underscores
		convertedInner := strings.ReplaceAll(inner, " ", "_")

		// Add the dot prefix
		return "{{ ." + convertedInner + " }}"
	})

	return result
}
