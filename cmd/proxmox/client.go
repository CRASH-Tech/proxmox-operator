package proxmox

import (
	"errors"
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
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

func (client *Client) getClusterApiConfig(clusterName string) (common.ApiConfig, error) {
	clusterApiConfig, isExists := client.Clusters[clusterName]
	if !isExists {
		return clusterApiConfig, errors.New(fmt.Sprintf("Unknown cluster: %s", clusterName))
	}
	return client.Clusters[clusterName], nil
}

func (client *Client) QemuCreate(cluster string, qemuConfig qemu.QemuConfig) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		qemu.Create(clusterApiConfig, qemuConfig)
	} else {
		fmt.Println(err)
	}
}
