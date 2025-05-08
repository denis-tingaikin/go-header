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

package main

import (
	"flag"
	"fmt"
	"os"

	"log"

	goheader "github.com/denis-tingaikin/go-header"
	"github.com/denis-tingaikin/go-header/version"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var defaultConfigPath = ".go-header.yml"

var flagSet flag.FlagSet

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		log.Fatal("Paths has not passed")
	}
	if len(paths) == 1 {
		if paths[0] == "version" {
			fmt.Println(version.Value())
			return
		}
	}

	var configPath string
	flagSet.StringVar(&configPath, "config", defaultConfigPath, "Path to config file")
	var analyser = goheader.NewAnalyzer(&configPath)
	analyser.Flags = flagSet

	singlechecker.Main(analyser)
}
