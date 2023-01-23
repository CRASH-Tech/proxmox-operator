package proxmox

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type StorageConfig struct {
	Node     string `json:"node"`
	VmId     int    `json:"vmid"`
	Filename string `json:"filename"`
	Size     string `json:"size"`
	Storage  string `json:"storage"`
}

func (node *Node) StorageCreate(storageConfig StorageConfig) error {
	log.Infof("Creating storage, cluster: %s, node: %s config: %+v", node.cluster.name, node.name, storageConfig)
	apiPath := fmt.Sprintf("/nodes/%s/storage/%s/content", node.name, storageConfig.Storage)
	err := node.cluster.PostReq(apiPath, storageConfig)
	if err != nil {
		return err
	}

	return nil
}
