/*
Copyright 2024 The CRASH-Tech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"io/ioutil"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"gopkg.in/yaml.v2"
)

func readConfig(path string) (common.Config, error) {
	config := common.Config{}
	config.Clusters = make(map[string]proxmox.ClusterApiConfig)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return common.Config{}, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return common.Config{}, err
	}

	return config, err
}
