package proxmox

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
	log "github.com/sirupsen/logrus"
)

func checkQemuConfig(qemuConfig qemu.QemuConfig) error {
	if _, isExist := qemuConfig["node"]; !isExist {
		return fmt.Errorf("no node name in qemu config")
	}
	if _, isExist := qemuConfig["vmid"]; !isExist {
		return fmt.Errorf("no vmid in qemu config")
	}

	return nil
}

func (client *Client) QemuCreate(cluster string, qemuConfig qemu.QemuConfig) {
	if err := checkQemuConfig(qemuConfig); err != nil {
		log.Error(err)
		return
	}
	if clusterApiConfig, err := client.getApiConfig(cluster); err == nil {
		log.Infof("Creating qemu VM: %s: %+v", cluster, qemuConfig)
		if err := qemu.Create(clusterApiConfig, qemuConfig); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuSetConfig(cluster string, qemuConfig qemu.QemuConfig) {
	if err := checkQemuConfig(qemuConfig); err != nil {
		log.Error(err)
		return
	}
	if clusterApiConfig, err := client.getApiConfig(cluster); err == nil {
		log.Infof("Set qemu VM config: %s: %+v", cluster, qemuConfig)
		if err := qemu.SetConfig(clusterApiConfig, qemuConfig); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuDelete(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getApiConfig(cluster); err == nil {
		log.Infof("Deleting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Delete(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuStart(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getApiConfig(cluster); err == nil {
		log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Start(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}

func (client *Client) QemuStop(cluster, node string, vmId int) {
	if clusterApiConfig, err := client.getApiConfig(cluster); err == nil {
		log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
		if err := qemu.Stop(clusterApiConfig, node, vmId); err != nil {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}
}
