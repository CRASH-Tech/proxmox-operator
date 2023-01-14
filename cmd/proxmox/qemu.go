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

func (client *Client) QemuCreate(cluster string, qemuConfig qemu.QemuConfig) error {
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Creating qemu VM, cluster: %s config: %+v", cluster, qemuConfig)
	err = qemu.Create(apiConfig, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) QemuSetConfig(cluster string, qemuConfig qemu.QemuConfig) error {
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Set qemu VM config, cluster: %s config: %+v", cluster, qemuConfig)
	err = qemu.SetConfig(apiConfig, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) QemuGetConfig(cluster, node string, vmId int) (qemu.QemuConfig, error) {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return nil, err
	}

	log.Infof("Get qemu VM config, cluster: %s node: %s vmid: %d", cluster, node, vmId)
	qemuConfig, err := qemu.GetConfig(apiConfig, node, vmId)
	if err != nil {
		return nil, err
	}

	return qemuConfig, nil
}

func (client *Client) QemuDelete(cluster, node string, vmId int) error {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Deleting qemu VM, cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Delete(apiConfig, node, vmId)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) QemuStart(cluster, node string, vmId int) error {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Starting qemu VM, cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Start(apiConfig, node, vmId)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) QemuStop(cluster, node string, vmId int) error {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		return err
	}

	log.Infof("Starting qemu VM, cluster: %s node: %s vmid: %d", cluster, node, vmId)
	err = qemu.Stop(apiConfig, node, vmId)
	if err != nil {
		return err
	}

	return nil
}
