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
	"testing"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstValue(t *testing.T) {
	t.Run("Get returns RawValue before Calculate", func(t *testing.T) {
		v := &goheader.ConstValue{RawValue: "hello"}
		assert.Equal(t, "hello", v.Get())
	})

	t.Run("Get returns calculated value after Calculate", func(t *testing.T) {
		v := &goheader.ConstValue{RawValue: "year-{{.YEAR}}"}
		err := v.Calculate(map[string]goheader.Value{
			"YEAR": &goheader.ConstValue{RawValue: "2025"},
		})
		require.NoError(t, err)
		assert.Equal(t, "year-2025", v.Get())
	})

	t.Run("String returns same as Get", func(t *testing.T) {
		v := &goheader.ConstValue{RawValue: "test"}
		assert.Equal(t, v.Get(), v.String())
	})

	t.Run("Raw returns RawValue", func(t *testing.T) {
		v := &goheader.ConstValue{RawValue: "raw"}
		assert.Equal(t, "raw", v.Raw())
	})

	t.Run("Clone is independent copy", func(t *testing.T) {
		original := &goheader.ConstValue{RawValue: "raw", Value: "val"}
		cloned := original.Clone().(*goheader.ConstValue)
		assert.Equal(t, "raw", cloned.RawValue)
		assert.Equal(t, "val", cloned.Value)

		cloned.RawValue = "modified"
		assert.Equal(t, "raw", original.RawValue)
	})
}

func TestRegexpValue(t *testing.T) {
	t.Run("Get returns RawValue before Calculate", func(t *testing.T) {
		v := &goheader.RegexpValue{RawValue: `[a-z]+`}
		assert.Equal(t, `[a-z]+`, v.Get())
	})

	t.Run("Get returns calculated value after Calculate", func(t *testing.T) {
		v := &goheader.RegexpValue{RawValue: "{{.PREFIX}}-[0-9]+"}
		err := v.Calculate(map[string]goheader.Value{
			"PREFIX": &goheader.ConstValue{RawValue: "ID"},
		})
		require.NoError(t, err)
		assert.Equal(t, "ID-[0-9]+", v.Get())
	})

	t.Run("String returns same as Get", func(t *testing.T) {
		v := &goheader.RegexpValue{RawValue: "test"}
		assert.Equal(t, v.Get(), v.String())
	})

	t.Run("Raw returns RawValue", func(t *testing.T) {
		v := &goheader.RegexpValue{RawValue: "raw"}
		assert.Equal(t, "raw", v.Raw())
	})

	t.Run("Clone is independent copy", func(t *testing.T) {
		original := &goheader.RegexpValue{RawValue: "raw", Value: "val"}
		cloned := original.Clone().(*goheader.RegexpValue)
		assert.Equal(t, "raw", cloned.RawValue)
		assert.Equal(t, "val", cloned.Value)

		cloned.RawValue = "modified"
		assert.Equal(t, "raw", original.RawValue)
	})
}

func TestCalculateValue(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		vals     map[string]goheader.Value
		expected string
	}{
		{
			name:     "no references",
			raw:      "plain text",
			vals:     map[string]goheader.Value{},
			expected: "plain text",
		},
		{
			name: "single reference",
			raw:  "hello {{.NAME}}",
			vals: map[string]goheader.Value{
				"NAME": &goheader.ConstValue{RawValue: "world"},
			},
			expected: "hello world",
		},
		{
			name: "reference with dot prefix and spaces",
			raw:  "{{ .YEAR }}",
			vals: map[string]goheader.Value{
				"YEAR": &goheader.ConstValue{RawValue: "2025"},
			},
			expected: "2025",
		},
		{
			name: "multiple references in one string",
			raw:  "{{.A}} and {{.B}}",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "first"},
				"B": &goheader.ConstValue{RawValue: "second"},
			},
			expected: "first and second",
		},
		{
			name: "adjacent references",
			raw:  "{{.A}}{{.B}}",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "hello"},
				"B": &goheader.ConstValue{RawValue: "world"},
			},
			expected: "helloworld",
		},
		{
			name: "reference at start",
			raw:  "{{.A}}-suffix",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "prefix"},
			},
			expected: "prefix-suffix",
		},
		{
			name: "reference at end",
			raw:  "prefix-{{.A}}",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "suffix"},
			},
			expected: "prefix-suffix",
		},
		{
			name: "nested reference",
			raw:  "{{.A}}",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "{{.B}}-end"},
				"B": &goheader.ConstValue{RawValue: "deep"},
			},
			expected: "deep-end",
		},
		{
			name:     "empty string",
			raw:      "",
			vals:     map[string]goheader.Value{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &goheader.ConstValue{RawValue: tt.raw}
			err := v.Calculate(tt.vals)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, v.Get())
		})
	}
}

func TestCalculateValue_Errors(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		vals    map[string]goheader.Value
		errMsg  string
	}{
		{
			name:   "missing closing braces",
			raw:    "hello {{.FOO",
			vals:   map[string]goheader.Value{},
			errMsg: "missed value ending",
		},
		{
			name:   "unknown value reference",
			raw:    "hello {{.UNKNOWN}}",
			vals:   map[string]goheader.Value{},
			errMsg: "unknown value name",
		},
		{
			name: "error in nested reference",
			raw:  "{{.A}}",
			vals: map[string]goheader.Value{
				"A": &goheader.ConstValue{RawValue: "{{.MISSING}}"},
			},
			errMsg: "unknown value name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &goheader.ConstValue{RawValue: tt.raw}
			err := v.Calculate(tt.vals)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}

	// Same errors apply to RegexpValue
	for _, tt := range tests {
		t.Run("regexp/"+tt.name, func(t *testing.T) {
			v := &goheader.RegexpValue{RawValue: tt.raw}
			err := v.Calculate(tt.vals)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}
