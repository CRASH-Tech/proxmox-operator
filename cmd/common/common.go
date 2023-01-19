package common

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Clusters      map[string]common.ApiConfig `yaml:"clusters"`
	DynamicClient *dynamic.DynamicClient
}
