package proxmox

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
	log "github.com/sirupsen/logrus"
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
		return clusterApiConfig, fmt.Errorf("unknown cluster: %s", clusterName)
	}
	return client.Clusters[clusterName], nil
}

func (client *Client) QemuCreate(cluster string, qemuConfig qemu.QemuConfig) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		log.Infof("Creating qemu VM: %s: %+v", cluster, qemuConfig)
		if err := qemu.Create(clusterApiConfig, qemuConfig); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuDelete(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		log.Infof("Deleting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Delete(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuStart(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Start(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuStop(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Stop(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}
