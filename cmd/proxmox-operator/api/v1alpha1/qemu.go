package v1alpha1 нужно как-то обрабатывать разные версии, возможно это пакет qemu, а не v1alpha1

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

func Get(api api.Api, resourceId schema.GroupVersionResource, name string) (Qemu, error) {
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

func GetAll(api api.Api, resourceId schema.GroupVersionResource) ([]Qemu, error) {
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
