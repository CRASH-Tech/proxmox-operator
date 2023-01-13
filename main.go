package main

import (
	"flag"
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
	TestSetVMConfig(client)
	//TestDeleteVM(client)
	//TestStartVM(client)
	//TestStopVM(client)
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
	qemuConfig := qemu.QemuConfig{
		VMId:    222,
		Node:    "crash-lab",
		OsType:  "l26",
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
		Name:    "k-test-c-33",
		Hotplug: "network,disk,usb",
		Agent:   "0",
		Numa:    1,
		Memory:  8192,
		Net0:    "virtio=A2:7B:45:48:9C:E6,bridge=vmbr0,tag=103",
		Cores:   8,
		Tablet:  1,
	}
	client.QemuCreate("crash-lab", qemuConfig)
}

func TestSetVMConfig(client *proxmox.Client) {
	qemuConfig := qemu.QemuConfig{
		VMId:    222,
		Node:    "crash-lab",
		OsType:  "l26",
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
		Name:    "k-test-c-33",
		Hotplug: "network,disk,usb",
		Agent:   "0",
		Numa:    1,
		Memory:  8192,
		Net0:    "virtio=A2:7B:45:48:9C:E6,bridge=vmbr0,tag=103",
		Cores:   8,
		Tablet:  1,
	}
	client.QemuSetConfig("crash-lab", "crash-lab", 222, qemuConfig)
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
