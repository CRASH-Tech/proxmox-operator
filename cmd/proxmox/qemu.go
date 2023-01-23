package proxmox

import (
	"encoding/json"
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	log "github.com/sirupsen/logrus"
)

type Qemu struct {
	node *Node
}

type (
	QemuConfig map[string]interface{}
)

func checkQemuConfig(qemuConfig QemuConfig) error {
	if _, isExist := qemuConfig["node"]; !isExist {
		return fmt.Errorf("no node name in qemu config")
	}
	if _, isExist := qemuConfig["vmid"]; !isExist {
		return fmt.Errorf("no vmid in qemu config")
	}

	return nil
}

func (qemu *Qemu) Create(qemuConfig QemuConfig) error {
	log.Infof("Creating qemu VM, cluster: %s, node: %s config: %+v", qemu.node.cluster.name, qemu.node.name, qemuConfig)
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("/nodes/%s/qemu", qemu.node.name)
	err = common.PostReq(qemu.node.cluster.apiConfig, apiPath, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) SetConfig(qemuConfig QemuConfig) error {
	log.Infof("Set qemu VM config, cluster: %s, node: %s config: %+v", qemu.node.cluster.name, qemu.node.name, qemuConfig)
	err := checkQemuConfig(qemuConfig)
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/config", qemu.node.name, qemuConfig["vmid"])
	err = common.PostReq(qemu.node.cluster.apiConfig, apiPath, qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) GetConfig(vmId int) (QemuConfig, error) {
	log.Infof("Get qemu VM config, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/config", qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)
	qemuConfigData, err := common.GetReq(qemu.node.cluster.apiConfig, apiPath, data)
	if err != nil {
		return nil, err
	}

	qemuConfig := QemuConfig{}
	err = json.Unmarshal(qemuConfigData, &qemuConfig)
	if err != nil {
		return nil, err
	}

	return qemuConfig, nil
}

func (qemu *Qemu) Delete(vmId int) error {
	log.Infof("Deleting qemu VM, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d", qemu.node.name, vmId)
	err := common.DeleteReq(qemu.node.cluster.apiConfig, apiPath)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) Start(vmId int) error {
	log.Infof("Starting qemu VM, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", qemu.node.name, vmId)
	err := common.PostReq(qemu.node.cluster.apiConfig, apiPath, data)
	if err != nil {
		return err
	}

	return nil
}

func (qemu *Qemu) Stop(vmId int) error {
	log.Infof("Starting qemu VM, cluster: %s node: %s vmid: %d", qemu.node.cluster.name, qemu.node.name, vmId)
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, qemu.node.name, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", qemu.node.name, vmId)
	err := common.PostReq(qemu.node.cluster.apiConfig, apiPath, data)
	if err != nil {
		return err
	}

	return nil
}
