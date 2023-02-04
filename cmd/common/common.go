package common

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Log           LogConfig                           `yaml:"log"`
	Clusters      map[string]proxmox.ClusterApiConfig `yaml:"clusters"`
	DynamicClient *dynamic.DynamicClient
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
