package common

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Debug         int                                 `yaml:"debug"`
	Clusters      map[string]proxmox.ClusterApiConfig `yaml:"clusters"`
	DynamicClient *dynamic.DynamicClient
}
