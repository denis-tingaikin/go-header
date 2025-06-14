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

	goheader "github.com/denis-tingaikin/go-header"
	"golang.org/x/tools/go/analysis/singlechecker"
)

const defaultConfigPath = ".go-header.yml"

func main() {
	cfgFlags := &ConfigFlag{
		configPath: defaultConfigPath,
		settings:   &goheader.Settings{},
	}

	var flagSet flag.FlagSet

	flagSet.Var(cfgFlags, "config", "path to the configuration file")

	analyser := goheader.New(cfgFlags.settings)

	analyser.Flags = flagSet

	singlechecker.Main(analyser)
}

type ConfigFlag struct {
	configPath string

	settings *goheader.Settings
}

func (c ConfigFlag) String() string {
	if len(c.configPath) != 0 {
		// Ignore errors because `String` is called before `Set`.
		cfg, _ := goheader.Parse(c.configPath)
		if cfg != nil {
			_ = cfg.FillSettings(c.settings)
		}
	}

	return c.configPath
}

func (c ConfigFlag) Set(w string) error {
	if w == "" {
		w = defaultConfigPath
	}

	c.configPath = w

	if len(c.configPath) != 0 {
		cfg, err := goheader.Parse(c.configPath)
		if err != nil {
			return err
		}

		err = cfg.FillSettings(c.settings)
		if err != nil {
			return err
		}
	}

	return nil
}
