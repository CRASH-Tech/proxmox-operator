package qemu

import (
	"fmt"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

// {"data":
// {"ostype":"l26",
// "bios":"seabios",
// "digest":"a2019681a11059dc5602d474c3443b907db213b7",
// "onboot":0,
// "smbios1":"uuid=3ae878b3-a77e-4a4a-adc6-14ee88350d36,manufacturer=MTIz,product=MTIz,version=MTIz,serial=MTIz,sku=MTIz,family=MTIz,base64=1",
// "scsi0":"local-lvm:vm-107-disk-0,size=32G",
// "sockets":1,
// "scsihw":"virtio-scsi-pci",
// "boot":"order=net0;ide2;scsi0",
// "cpu":"host",
// "meta":"creation-qemu=6.2.0,ctime=1673361689",
// "ide2":"none,media=cdrom",
// "kvm":1,
// "name":"k-test-c-3",
// "hotplug":"network,disk,usb",
// "agent":"0",
// "numa":1,
// "memory":8192,
// "net0":"virtio=A2:7B:45:48:9C:E5,bridge=vmbr0,tag=103",
// "cores":8,
// "tablet":1,
// "vmgenid":"10e77c08-c74f-466e-9186-eb5f8da4079c"}}
type QemuConfig struct {
	VMId    int    `json:"vmid,omitempty"`
	Node    string `json:"node,omitempty"`
	OsType  string `json:"ostype,omitempty"`
	Bios    string `json:"bios,omitempty"`
	Digest  string `json:"digest,omitempty"`
	Onboot  int    `json:"onboot,omitempty"`
	Smbios1 string `json:"smbios1,omitempty"`
	Scsi0   string `json:"scsi0,omitempty"`
	Sockets int    `json:"sockets,omitempty"`
	Scsihw  string `json:"scsihw,omitempty"`
	Boot    string `json:"boot,omitempty"`
	CPU     string `json:"cpu,omitempty"`
	Meta    string `json:"meta,omitempty"`
	Ide2    string `json:"ide2,omitempty"`
	Kvm     int    `json:"kvm,omitempty"`
	Name    string `json:"name,omitempty"`
	Hotplug string `json:"hotplug,omitempty"`
	Agent   string `json:"agent,omitempty"`
	Numa    int    `json:"numa,omitempty"`
	Memory  int    `json:"memory,omitempty"`
	Net0    string `json:"net0,omitempty"`
	Cores   int    `json:"cores,omitempty"`
	Tablet  int    `json:"tablet,omitempty"`
	Vmgenid string `json:"vmgenid,omitempty"`
}

func Create(clusterApiConfig common.ApiConfig, qemuConfig QemuConfig) error {
	apiPath := fmt.Sprintf("/nodes/%s/qemu", qemuConfig.Node)
	err := common.PostReq(clusterApiConfig, apiPath, qemuConfig)

	return err
}

func Delete(clusterApiConfig common.ApiConfig, node string, vmId int) error {
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d", node, vmId)
	err := common.DeleteReq(clusterApiConfig, apiPath)

	return err
}

func Start(clusterApiConfig common.ApiConfig, node string, vmId int) error {
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, node, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/start", node, vmId)
	err := common.PostReq(clusterApiConfig, apiPath, data)

	return err
}

func Stop(clusterApiConfig common.ApiConfig, node string, vmId int) error {
	data := fmt.Sprintf(`{"node":"%s", "vmid":"%d"}`, node, vmId)
	apiPath := fmt.Sprintf("/nodes/%s/qemu/%d/status/stop", node, vmId)
	err := common.PostReq(clusterApiConfig, apiPath, data)

	return err
}
