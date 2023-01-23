package v1alpha1

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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

var (
	resourceId = schema.GroupVersionResource{
		Group:    "proxmox.xfix.org",
		Version:  "v1alpha1",
		Resource: "qemu",
	}
)

func QemuGet(api api.Api, name string) (Qemu, error) {
	item, err := api.DynamicGetClusterResource(api.Ctx, &api.Dynamic, resourceId, name)
	if err != nil {
		panic(err)
	}

	var qemu Qemu
	err = json.Unmarshal(item, &qemu)
	if err != nil {
		return Qemu{}, err
	}

	return qemu, nil
}

func QemuGetAll(api api.Api) ([]Qemu, error) {
	items, err := api.DynamicGetClusterResources(api.Ctx, &api.Dynamic, resourceId)
	if err != nil {
		panic(err)
	}

	var result []Qemu
	for _, item := range items {
		var qemu Qemu
		err = json.Unmarshal(item, &qemu)
		if err != nil {
			return nil, err
		}
		result = append(result, qemu)
	}

	return result, nil
}

func QemuPatch(api api.Api, qemu Qemu) error {
	jsonData, err := json.Marshal(qemu)
	if err != nil {
		return err
	}

	_, err = api.DynamicPatchClusterResource(api.Ctx, &api.Dynamic, resourceId, qemu.Metadata.Name, jsonData)
	if err != nil {
		return err
	}

	return nil
}
