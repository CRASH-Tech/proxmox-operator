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
