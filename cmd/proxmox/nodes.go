package proxmox

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes"
	log "github.com/sirupsen/logrus"
)

func (client *Client) NodesGet(cluster string) ([]nodes.Node, error) {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return nil, err
	}

	log.Infof("Get cluster nodes: %s", cluster)
	nodes, err := nodes.Get(apiConfig)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
