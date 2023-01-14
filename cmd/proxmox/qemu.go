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
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		log.Error(err)
		return
	}

	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Creating qemu VM: %s: %+v", cluster, qemuConfig)
	err = qemu.Create(apiConfig, qemuConfig)
	if err != nil {
		log.Error(err)
		return
	}

}

func (client *Client) QemuSetConfig(cluster string, qemuConfig qemu.QemuConfig) {
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		log.Error(err)
		return
	}

	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Set qemu VM config: %s: %+v", cluster, qemuConfig)
	err = qemu.SetConfig(apiConfig, qemuConfig)
	if err != nil {
		log.Error(err)
		return
	}

}

func (client *Client) QemuDelete(cluster, node string, vmId int) {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Deleting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Delete(apiConfig, node, vmId)
	if err != nil {
		log.Error(err)
		return
	}

}

func (client *Client) QemuStart(cluster, node string, vmId int) {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Start(apiConfig, node, vmId)
	if err != nil {
		log.Error(err)
		return
	}

}

func (client *Client) QemuStop(cluster, node string, vmId int) {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infof("Starting qemu VM: cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Stop(apiConfig, node, vmId)
	if err != nil {
		log.Error(err)
		return
	}

}
