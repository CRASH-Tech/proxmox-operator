package proxmox

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/cluster"
	log "github.com/sirupsen/logrus"
)

func (client *Client) ClusterGetNextId(c string) (int, error) {
	apiConfig, err := client.getApiConfig(c)
	if err != nil {
		return -1, err
	}

	log.Infof("Get next id: %s", c)
	nextId, err := cluster.GetNextId(apiConfig)
	if err != nil {
		return -1, err
	}

	return nextId, err
}
