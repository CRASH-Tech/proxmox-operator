package v1alpha1

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Qemu struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   *Metadata `json:"metadata"`
	Spec       *Spec     `json:"spec"`
}
type Metadata struct {
	Name string `json:"name"`
}
type Config struct {
	Agent   bool   `json:"agent"`
	Cores   int    `json:"cores"`
	Sockets int    `json:"sockets"`
	Test    string `json:"test"`
}
type Spec struct {
	Accepted bool    `json:"accepted"`
	Cluster  string  `json:"cluster"`
	Config   *Config `json:"config"`
	Node     string  `json:"node"`
	Pool     string  `json:"pool"`
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
