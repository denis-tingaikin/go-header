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
	"log"

	goheader "github.com/denis-tingaikin/go-header"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var defaultConfigPath = ".go-header.yml"

var flagSet flag.FlagSet

func main() {
	var configPath string

	flagSet.StringVar(&configPath, "config", defaultConfigPath, "path to config file")

	var cfg goheader.Config
	if err := cfg.Parse(configPath); err != nil {
		log.Fatal(err)
	}

	analyser, err := goheader.NewAnalyzer(&cfg)
	analyser.Flags = flagSet

	if err != nil {
		log.Fatal(err)
	}

	singlechecker.Main(analyser)
}
