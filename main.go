package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Clusters map[string]common.ApiConfig `yaml:"clusters"`
}

var (
	version = "0.0.1"
	config  Config
)

func init() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "config file path. Default: config.yaml")
	config = getConfig(configPath)
	//fmt.Println(config)
}

func main() {
	log.Infof("Starting proxmox-operator %s\n", version)
	client := proxmox.NewClient(config.Clusters)

	//TestCreateVM(client)
	//TestSetVMConfig(client)
	//TestStartVM(client)
	//TestStopVM(client)
	//TestDeleteVM(client)
	TestGetNodes(client)
}

func getConfig(path string) (result Config) {
	result.Clusters = make(map[string]common.ApiConfig)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read config file: %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &result)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return
}

func TestCreateVM(client *proxmox.Client) {
	qemuConfig := qemu.QemuConfig{}

	qemuConfig["vmid"] = 222
	qemuConfig["node"] = "crash-lab"
	qemuConfig["name"] = "k-test-c-33"
	qemuConfig["ostype"] = "l26"
	qemuConfig["bios"] = "seabios"
	qemuConfig["onboot"] = 0
	qemuConfig["smbios1"] = "uuid=3ae878b3-a77e-4a4a-adc6-14ee88350d36,manufacturer=MTIz,product=MTIz,version=MTIz,serial=MTIz,sku=MTIz,family=MTIz,base64=1"
	qemuConfig["scsi0"] = "local-lvm:vm-107-disk-0,size=32G"
	qemuConfig["sockets"] = 1
	qemuConfig["scsihw"] = "virtio-scsi-pci"
	qemuConfig["boot"] = "order=net0;ide2;scsi0"
	qemuConfig["cpu"] = "host"
	qemuConfig["ide2"] = "none,media=cdrom"
	qemuConfig["kvm"] = 1
	qemuConfig["hotplug"] = "network,disk,usb"
	qemuConfig["agent"] = "0"
	qemuConfig["numa"] = 1
	qemuConfig["memory"] = 8192
	qemuConfig["net0"] = "virtio=A2:7B:45:48:9C:E6,bridge=vmbr0,tag=103"
	qemuConfig["cores"] = 8
	qemuConfig["tablet"] = 1

	client.QemuCreate("crash-lab", qemuConfig)
}

func TestSetVMConfig(client *proxmox.Client) {
	qemuConfig := qemu.QemuConfig{}

	qemuConfig["vmid"] = 222
	qemuConfig["node"] = "crash-lab"
	qemuConfig["name"] = "k-test-c-44"
	qemuConfig["node"] = "crash-lab"
	qemuConfig["ostype"] = "l26"
	qemuConfig["bios"] = "seabios"
	qemuConfig["onboot"] = 0
	qemuConfig["smbios1"] = "uuid=3ae878b3-a77e-4a4a-adc6-14ee88350d36,manufacturer=MTIz,product=MTIz,version=MTIz,serial=MTIz,sku=MTIz,family=MTIz,base64=1"
	qemuConfig["scsi0"] = "local-lvm:vm-107-disk-0,size=32G"
	qemuConfig["sockets"] = 1
	qemuConfig["scsihw"] = "virtio-scsi-pci"
	qemuConfig["boot"] = "order=net0;ide2;scsi0"
	qemuConfig["cpu"] = "host"
	qemuConfig["ide2"] = "none,media=cdrom"
	qemuConfig["kvm"] = 1
	qemuConfig["hotplug"] = "network,disk,usb"
	qemuConfig["agent"] = "0"
	qemuConfig["numa"] = 1
	qemuConfig["memory"] = 8192
	qemuConfig["net0"] = "virtio=A2:7B:45:48:9C:E6,bridge=vmbr0,tag=103"
	qemuConfig["cores"] = 8
	qemuConfig["tablet"] = 1

	client.QemuSetConfig("crash-lab", qemuConfig)
}

func TestDeleteVM(client *proxmox.Client) {
	client.QemuDelete("crash-lab", "crash-lab", 222)
}

func TestStartVM(client *proxmox.Client) {
	client.QemuStart("crash-lab", "crash-lab", 222)
}

func TestStopVM(client *proxmox.Client) {
	client.QemuStop("crash-lab", "crash-lab", 222)
}

func TestGetNodes(client *proxmox.Client) {
	nodes, err := client.NodesGet("crash-lab")
	if err != nil {
		log.Error(err)
		return
	}

	for _, node := range nodes {
		fmt.Println(node.ID)
	}
}
