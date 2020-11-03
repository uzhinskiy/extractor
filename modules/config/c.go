// Copyright © 2020 Uzhinskiy Boris
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	App struct {
		Port string `yaml:"port"`
	} `yaml:"app"`
	Elastic struct {
		Hosts []string `yaml:"hosts"`
		SSL   bool     `yaml:"ssl"`
		Cert  string   `yaml:"certfile"`
	} `yaml:"elastic"`
}

func Parse(f string) Config {
	var c Config
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		panic(err)
	}

	if c.App.Port == "" {
		c.App.Port = "9400"
	}

	if len(c.Elastic.Hosts) == 0 {
		c.Elastic.Hosts[0] = "http://127.0.0.1:9200/"
	}

	return c
}
