package proxmox

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DiskConfig struct {
	Name     string `json:"-"`
	Node     string `json:"node"`
	VmId     int    `json:"vmid"`
	Filename string `json:"filename"`
	Size     string `json:"size"`
	Storage  string `json:"storage"`
}

func (node *Node) DiskCreate(diskConfig DiskConfig) error {
	log.Debugf("Creating disk, cluster: %s, node: %s config: %+v", node.cluster.name, node.name, diskConfig)
	apiPath := fmt.Sprintf("/nodes/%s/storage/%s/content", node.name, diskConfig.Storage)
	err := node.cluster.PostReq(apiPath, diskConfig)
	if err != nil {
		return err
	}

	return nil
}
