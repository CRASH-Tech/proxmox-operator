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

	log.Infof("Get next id, cluster: %s", c)
	nextId, err := cluster.GetNextId(apiConfig)
	if err != nil {
		return -1, err
	}

	return nextId, err
}

func (client *Client) ClusterGetResources(c string) ([]cluster.Resource, error) {
	apiConfig, err := client.getApiConfig(c)
	if err != nil {
		return nil, err
	}

	log.Infof("Get cluster resources, cluster: %s", c)
	resources, err := cluster.GetResources(apiConfig)
	if err != nil {
		return nil, err
	}

	return resources, err
}
