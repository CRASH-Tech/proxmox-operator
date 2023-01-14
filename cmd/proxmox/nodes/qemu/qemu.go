package qemu

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type (
	QemuConfig map[string]interface{}
)

func Create(apiConfig common.ApiConfig, qemuConfig QemuConfig) error {
	apiPath := fmt.Sprintf("/nodes/%s/qemu", qemuConfig["node"])
	err := common.PostReq(apiConfig, apiPath, qemuConfig)

	return err
}

func SetConfig(apiConfig common.ApiConfig, qemuConfig QemuConfig) error {
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/config", qemuConfig["node"], qemuConfig["vmid"])
	err := common.PostReq(apiConfig, apiPath, qemuConfig)

	return err
}

func Delete(apiConfig common.ApiConfig, node string, vmId int) error {
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d", node, vmId)
	err := common.DeleteReq(apiConfig, apiPath)

	return err
}

func Start(apiConfig common.ApiConfig, node string, vmId int) error {
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, node, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", node, vmId)
	err := common.PostReq(apiConfig, apiPath, data)

	return err
}

func Stop(apiConfig common.ApiConfig, node string, vmId int) error {
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, node, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", node, vmId)
	err := common.PostReq(apiConfig, apiPath, data)

	return err
}
