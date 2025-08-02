# go-header
[![ci](https://github.com/denis-tingaikin/go-header/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/denis-tingaikin/go-header/actions/workflows/ci.yml)

Simple go source code linter providing checks for copyrgiht headers. 

## Features

| Feature                     | Status | Details                                  |
|-----------------------------|--------|------------------------------------------|
| ✅ **Copyright Headers**     | ✔️     | Supports all standard formats            |
| ✅ **Parallel Processing**   | ✔️     | Processes files concurrently             |
| ✅ **Comment Support**       | ✔️     | `//`, `/* */`, `/* * */`                 |
| ✅ **Go/Analysis**           | ✔️     | Native Go tooling integration            |
| ✅ **Regex Customization**   | ✔️     | User-defined pattern matching            |
| ✅ **Automatic Year Checks** | ✔️     | Validates & updates copyright years      |
| ✅ **Auto-Fix Files**        | ✔️     | In-place header corrections              |
| ✅ **Go/Template Support**   | ✔️     | go templates can be used in headers      |
| ✍️ **Multi-License Support** | ❌     | In progress                              |



## Installation

For installation you can simply use `go install`.

```bash
go install github.com/denis-tingaikin/go-header/cmd/go-header@latest
```
## Usage

```bash
  -V    print version and exit
  -all
        no effect (deprecated)
  -c int
        display offending line with this many lines of context (default -1)
  -config string
        path to config file (default ".go-header.yml")
  -cpuprofile string
        write CPU profile to this file
  -debug string
        debug flags, any subset of "fpstv"
  -diff
        with -fix, don't update the files, but print a unified diff
  -fix
        apply all suggested fixes
  -flags
        print analyzer flags in JSON
  -json
        emit JSON output
  -memprofile string
        write memory profile to this file
  -source
        no effect (deprecated)
  -tags string
        no effect (deprecated)
  -test
        indicates whether test files should be analyzed, too (default true)
  -trace string
        write trace log to this file
  -v    no effect (deprecated)
```
## Configuration
To configuring `.go-header.yml` linter you simply need to fill the next fields:

```yaml
---
template: # expects header template string.
template-path: # expects path to file with license header string. 
values: # expects `const` or `regexp` node with values where values is a map string to string.
  const:
    key1: value1 # const value just checks equality. Note `key1` should be used in template string as {{ key1 }} or {{ KEY1 }}.
  regexp:
    key2: value2 # regexp value just checks regex match. The value should be a valid regexp pattern. Note `key2` should be used in template string as {{ key2 }} or {{ KEY2 }}.
```

Where `values` also can be used recursively. Example:

```yaml
values:
  const:
    key1: "value" 
  regexp:
    key2: "{{key1}} value1" # Reads as regex pattern "value value1"
```

## Bult-in values

- **MOD_YEAR** - Returns the year when the file was modified.
- **MOD_YEAR-RANGE** - Returns a year-range where the range starts from the  year when the file was modified.
- **YEAR** - Expects current year. Example header value: `2020`.  Example of template using: `{{YEAR}}` or `{{year}}`.
- **YEAR-RANGE** - Expects any valid year interval or current year. Example header value: `2020` or `2000-2020`. Example of template using: `{{year-range}}` or `{{YEAR-RANGE}}`.

## Execution

`go-header` linter expects file paths on input. If you want to run `go-header` only on diff files, then you can use this command:

```bash
go-header ./...
```

## Setup example

### Step 1

Create configuration file  `.go-header.yml` in the root of project.

```yaml
---
vars:
  DOMAIN: sales|product
  MY_COMPANY: {{ .DOMAIN }}.mycompany.com
template: |
  {{ .MY_COMPANY }}
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
```

### Step 2 
Run `go-header ./...`
