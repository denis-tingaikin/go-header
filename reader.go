/*
Copyright (c) 2020-2025 Denis Tingaikin

SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package goheader

func NewReader(text string) *Reader {
	return &Reader{source: []rune(text)}
}

type Reader struct {
	source   []rune
	position int
	location Location
	offset   Location
}

func (r *Reader) SetOffset(offset Location) {
	r.offset = offset
}

func (r *Reader) Position() int {
	return r.position
}

func (r *Reader) Location() Location {
	return r.location.Add(r.offset)
}

func (r *Reader) Peek() rune {
	if r.Done() {
		return rune(0)
	}
	return r.source[r.position]
}

func (r *Reader) Done() bool {
	return r.position >= len(r.source)
}

func (r *Reader) Next() rune {
	if r.Done() {
		return rune(0)
	}
	result := r.Peek()
	if result == '\n' {
		r.location.Line++
		r.location.Position = 0
	} else {
		r.location.Position++
	}
	r.position++
	return result
}

func (r *Reader) Finish() string {
	if r.position >= len(r.source) {
		return ""
	}
	defer r.till()
	return string(r.source[r.position:])
}

func (r *Reader) SetPosition(pos int) {
	if pos < 0 {
		r.position = 0
	}
	r.position = pos
	r.location = r.calculateLocation()
}

func (r *Reader) ReadWhile(match func(rune) bool) string {
	if match == nil {
		return ""
	}
	start := r.position
	for !r.Done() && match(r.Peek()) {
		r.Next()
	}
	return string(r.source[start:r.position])
}

func (r *Reader) till() {
	r.position = len(r.source)
	r.location = r.calculateLocation()
}

func (r *Reader) calculateLocation() Location {
	minVal := min(len(r.source), r.position)
	x, y := 0, 0
	for i := 0; i < minVal; i++ {
		if r.source[i] == '\n' {
			y++
			x = 0
		} else {
			x++
		}
	}
	return Location{Line: y, Position: x}
}
