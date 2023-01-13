package qemu

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
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
type QemuVMConfigImpl struct {
	VMId    int    `json:"vmid,omitempty"`
	Node    string `json:"node,omitempty"`
	Ostype  string `json:"ostype,omitempty"`
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

//  root@pam!test
//  4f718e55-d6a4-4840-a548-193a6c84fece
//curl -H "Authorization: PVEAPIToken=root@pam!test=4f718e55-d6a4-4840-a548-193a6c84fece" https://10.0.0.1:8006/api2/json/
func Create() {
	PROXMOX_API_URL := "https://crash-lab.uis.st/api2/json/nodes/crash-lab/qemu"
	PROXMOX_API_TOKEN_ID := "root@pam!test"
	PROXMOX_API_TOKEN_SECRET := "4f718e55-d6a4-4840-a548-193a6c84fece"

	qemuConfig := QemuVMConfigImpl{
		VMId:    222,
		Node:    "crash-lab",
		Ostype:  "l26",
		Bios:    "seabios",
		Onboot:  0,
		Smbios1: "uuid=3ae878b3-a77e-4a4a-adc6-14ee88350d36,manufacturer=MTIz,product=MTIz,version=MTIz,serial=MTIz,sku=MTIz,family=MTIz,base64=1",
		Scsi0:   "local-lvm:vm-107-disk-0,size=32G",
		Sockets: 1,
		Scsihw:  "virtio-scsi-pci",
		Boot:    "order=net0;ide2;scsi0",
		CPU:     "host",
		Ide2:    "none,media=cdrom",
		Kvm:     1,
		Name:    "k-test-c-3",
		Hotplug: "network,disk,usb",
		Agent:   "0",
		Numa:    1,
		Memory:  8192,
		Net0:    "virtio=A2:7B:45:48:9C:E6,bridge=vmbr0,tag=103",
		Cores:   8,
		Tablet:  1,
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", PROXMOX_API_TOKEN_ID, PROXMOX_API_TOKEN_SECRET)).
		SetBody(qemuConfig).
		//		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		Post(PROXMOX_API_URL)
	fmt.Println(resp)
	fmt.Println(err)

	data, _ := json.Marshal(qemuConfig)
	fmt.Printf(string(data))
}
