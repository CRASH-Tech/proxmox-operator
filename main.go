package main

import (
	"fmt"
	"os"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/qemu"
)

var (
	proxmoxApiConfig proxmox.ProxmoxApiConfigImpl
	version          = "0.0.1"
)

func init() {
	proxmoxApiConfig.ApiUrl = os.Getenv("PROXMOX_API_URL")
	proxmoxApiConfig.ApiUsername = os.Getenv("PROXMOX_API_USERNAME")
	proxmoxApiConfig.ApiToken = os.Getenv("PROXMOX_API_TOKEN")
}

func main() {
	fmt.Printf("Starting proxmox-operator %s\n", version)
	qemu.Create()
}
