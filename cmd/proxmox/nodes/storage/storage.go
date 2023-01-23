package storage

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type StorageConfig struct {
	Node     string `json:"node"`
	VmId     int    `json:"vmid"`
	Filename string `json:"filename"`
	Size     string `json:"size"`
	Storage  string `json:"storage"`
}

func Create(apiConfig common.ApiConfig, storageConfig StorageConfig) error {
	apiPath := fmt.Sprintf("/nodes/%s/storage/%s/content", storageConfig.Node, storageConfig.Storage)
	err := common.PostReq(apiConfig, apiPath, storageConfig)

	return err
}
