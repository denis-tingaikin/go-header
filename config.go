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
	"errors"
	"fmt"
	"os"
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
}

func (c *Config) builtInValues() map[string]Value {
	var result = make(map[string]Value)
	year := fmt.Sprint(time.Now().Year())
	result["year-range"] = &RegexpValue{
		RawValue: `((20\d\d\-{{YEAR}})|({{YEAR}}))`,
	}
	result["year"] = &ConstValue{
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
			key := strings.ToLower(k)
			result[key] = create(v)
		}
	}
	appendValues(c.Values["const"], createConst)
	appendValues(c.Values["regexp"], createRegexp)
	appendValues(c.Vars, createRegexp)
	return result, nil
}

func (c *Config) GetTemplate() (string, error) {
	if c.Template != "" {
		return c.Template, nil
	}
	if c.TemplatePath == "" {
		return "", errors.New("template has not passed")
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
