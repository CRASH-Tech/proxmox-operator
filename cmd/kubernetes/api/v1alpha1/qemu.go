package v1alpha1

type Qemu struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   QemuMetadata `json:"metadata"`
	Spec       QemuSpec     `json:"spec"`
	Status     QemuStatus   `json:"status"`
}

type QemuMetadata struct {
	Name string `json:"name"`
}

type QemuSpec struct {
	Cluster string                 `json:"cluster"`
	Node    string                 `json:"node"`
	Pool    string                 `json:"pool"`
	Vmid    int                    `json:"vmid"`
	CPU     QemuCPU                `json:"cpu"`
	Memory  QemuMemory             `json:"memory"`
	Disk    []QemuDisk             `json:"disk"`
	Network []QemuNetwork          `json:"network"`
	Options map[string]interface{} `json:"options"`
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
	Status  string `json:"status"`
	Power   string `json:"power"`
	Cluster string `json:"cluster"`
	Node    string `json:"node"`
}
