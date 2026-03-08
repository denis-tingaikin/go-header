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
	"regexp"
	"strings"
)

type Value interface {
	Calculate(map[string]Value) error
	Get() string
	Raw() string
	Clone() Value
}

func calculateValue(calculable Value, values map[string]Value) (string, error) {
	return calculateValueWithCycleDetection(calculable, values, nil)
}

func calculateValueWithCycleDetection(calculable Value, values map[string]Value, seen map[string]bool) (string, error) {
	sb := strings.Builder{}
	r := calculable.Raw()
	for {
		startIndex := strings.Index(r, "{{")
		if startIndex < 0 {
			break
		}
		_, _ = sb.WriteString(r[:startIndex])
		endIndex := strings.Index(r[startIndex:], "}}")
		if endIndex < 0 {
			return "", errors.New("missed value ending")
		}
		endIndex += startIndex // convert to absolute index in r
		subVal := strings.TrimSpace(r[startIndex+2 : endIndex])
		subVal, _ = strings.CutPrefix(subVal, ".")
		if seen[subVal] {
			return "", fmt.Errorf("circular reference detected for value %v", subVal)
		}
		if val := values[subVal]; val != nil {
			nextSeen := make(map[string]bool, len(seen)+1)
			for k, v := range seen {
				nextSeen[k] = v
			}
			nextSeen[subVal] = true
			v, err := calculateValueWithCycleDetection(val, values, nextSeen)
			if err != nil {
				return "", err
			}
			sb.WriteString(v)
		} else {
			return "", fmt.Errorf("unknown value name %v", subVal)
		}
		r = r[endIndex+2:]
	}
	_, _ = sb.WriteString(r)
	return sb.String(), nil
}

type ConstValue struct {
	RawValue   string
	Value      string
	calculated bool
}

func (c *ConstValue) Calculate(values map[string]Value) error {
	v, err := calculateValue(c, values)
	if err != nil {
		return err
	}
	c.Value = v
	c.calculated = true
	return nil
}

func (c *ConstValue) Raw() string {
	return c.RawValue
}

func (c *ConstValue) Clone() Value {
	return &ConstValue{
		RawValue:   c.RawValue,
		Value:      c.Value,
		calculated: c.calculated,
	}
}

func (c *ConstValue) Get() string {
	if c.calculated {
		return c.Value
	}
	return c.RawValue
}

func (c *ConstValue) String() string {
	return c.Get()
}

type RegexpValue struct {
	RawValue   string
	Value      string
	calculated bool
}

func (r *RegexpValue) Clone() Value {
	return &RegexpValue{
		Value:      r.Value,
		RawValue:   r.RawValue,
		calculated: r.calculated,
	}
}

func (r *RegexpValue) Calculate(values map[string]Value) error {
	v, err := calculateValue(r, values)
	if err != nil {
		return err
	}
	r.Value = v
	r.calculated = true
	return nil
}

func (r *RegexpValue) Raw() string {
	return r.RawValue
}

func (r *RegexpValue) Get() string {
	if r.calculated {
		return r.Value
	}
	return r.RawValue
}

func (r *RegexpValue) String() string {
	return r.Get()
}

// regexSafeValue wraps a ConstValue so that String() returns
// a regex-escaped version of its value. This is used when substituting
// const values into the regex matching template.
type regexSafeValue struct {
	inner *ConstValue
}

func (r *regexSafeValue) String() string {
	return regexp.QuoteMeta(r.inner.Get())
}

// regexSafeValues returns a map where ConstValues are wrapped to produce
// regex-escaped output via String(). RegexpValues are left as-is since
// they intentionally contain regex patterns.
func regexSafeValues(vals map[string]Value) map[string]any {
	result := make(map[string]any, len(vals))
	for k, v := range vals {
		if cv, ok := v.(*ConstValue); ok {
			result[k] = &regexSafeValue{inner: cv}
		} else {
			result[k] = v
		}
	}
	return result
}

var _ Value = &ConstValue{}
var _ Value = &RegexpValue{}
