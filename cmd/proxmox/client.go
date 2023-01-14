package proxmox

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type Client struct {
	Clusters map[string]common.ApiConfig
}

func NewClient(clusters map[string]common.ApiConfig) *Client {
	client := Client{
		Clusters: clusters,
	}

	return &client
}

func (client *Client) getApiConfig(clusterName string) (common.ApiConfig, error) {
	clusterApiConfig, isExists := client.Clusters[clusterName]
	if !isExists {
		return clusterApiConfig, fmt.Errorf("unknown cluster: %s", clusterName)
	}
	return client.Clusters[clusterName], nil
}
