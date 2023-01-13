package proxmox

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
	log "github.com/sirupsen/logrus"
)

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

func (client *Client) QemuSetConfig(cluster string, node string, vmId int, qemuConfig qemu.QemuConfig) {
	if clusterApiConfig, err := client.getClusterApiConfig(cluster); err == nil {
		log.Infof("Set qemu VM config: %s: %+v", cluster, qemuConfig)
		if err := qemu.SetConfig(clusterApiConfig, node, vmId, qemuConfig); err != nil {
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
