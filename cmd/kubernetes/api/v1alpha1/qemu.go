package v1alpha1

import "github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes/api"

const (
	STATUS_QEMU_EMPTY       = ""
	STATUS_QEMU_SYNCED      = "SYNCED"
	STATUS_QEMU_OUT_OF_SYNC = "OUT OF SYNC"
	STATUS_QEMU_PENDING     = "PENDING"
	STATUS_QEMU_ERROR       = "ERROR"
	STATUS_QEMU_DELETING    = "DELETING"
	STATUS_QEMU_UNKNOWN     = "UNKNOWN"

	STATUS_POWER_ON      = "ON"
	STATUS_POWER_OFF     = "OFF"
	STATUS_POWER_UNKNOWN = "UNKNOWN"
)

type Qemu struct {
	*api.CustomResource
	Spec   QemuSpec   `json:"spec"`
	Status QemuStatus `json:"status"`
}

type QemuSpec struct {
	Autostart bool                   `json:"autostart"`
	Autostop  bool                   `json:"autostop"`
	Cluster   string                 `json:"cluster"`
	Node      string                 `json:"node"`
	Pool      string                 `json:"pool"`
	VmId      int                    `json:"vmid"`
	CPU       QemuCPU                `json:"cpu"`
	Memory    QemuMemory             `json:"memory"`
	Disk      []QemuDisk             `json:"disk"`
	Network   []QemuNetwork          `json:"network"`
	Options   map[string]interface{} `json:"options"`
}

type QemuCPU struct {
	Cores   int    `json:"cores"`
	Sockets int    `json:"sockets"`
	Type    string `json:"type"`
}

type QemuDisk struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	Storage string `json:"storage"`
}

type QemuMemory struct {
	Balloon int `json:"balloon"`
	Size    int `json:"size"`
}

type QemuNetwork struct {
	Bridge string `json:"bridge"`
	Mac    string `json:"mac"`
	Model  string `json:"model"`
	Name   string `json:"name"`
	Tag    int    `json:"tag"`
}

type QemuStatus struct {
	Deploy  string              `json:"deploy"`
	Power   string              `json:"power"`
	Cluster string              `json:"cluster"`
	Node    string              `json:"node"`
	VmId    int                 `json:"vmid"`
	Net     []QemuStatusNetwork `json:"net"`
}

type QemuStatusNetwork struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
}
