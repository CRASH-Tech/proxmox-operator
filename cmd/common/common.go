package common

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Clusters      map[string]common.ApiConfig `yaml:"clusters"`
	DynamicClient *dynamic.DynamicClient
}

func CrToStruct(data []byte, out interface{}) error {
	err := json.Unmarshal(data, out)
	if err != nil {
		return err
	}

	return nil
}
