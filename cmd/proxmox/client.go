package proxmox

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type Client struct {
	clusters map[string]common.ApiConfig
}

func NewClient(clusters map[string]common.ApiConfig) *Client {
	client := Client{
		clusters: clusters,
	}

	return &client
}

func (client *Client) getApiConfig(clusterName string) (common.ApiConfig, error) {
	clusterApiConfig, isExists := client.clusters[clusterName]
	if !isExists {
		return clusterApiConfig, fmt.Errorf("unknown cluster: %s", clusterName)
	}

	return client.clusters[clusterName], nil
}

func (client *Client) Cluster(cluster string) *Cluster {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	result := Cluster{
		name:      cluster,
		apiConfig: apiConfig,
	}

	return &result
}
