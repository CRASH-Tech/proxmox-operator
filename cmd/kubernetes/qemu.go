package kuberentes

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ()

type Qemu struct {
	client     *Client
	resourceId schema.GroupVersionResource
}

func (qemu *Qemu) Get(name string) (v1alpha1.Qemu, error) {
	item, err := qemu.client.dynamicGet(qemu.resourceId, name)
	if err != nil {
		panic(err)
	}

	var result v1alpha1.Qemu
	err = json.Unmarshal(item, &result)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	return result, nil
}

func (qemu *Qemu) GetAll() ([]v1alpha1.Qemu, error) {
	items, err := qemu.client.dynamicGetAll(qemu.resourceId)
	if err != nil {
		panic(err)
	}

	var result []v1alpha1.Qemu
	for _, item := range items {
		var q v1alpha1.Qemu
		err = json.Unmarshal(item, &q)
		if err != nil {
			return nil, err
		}

		result = append(result, q)
	}

	return result, nil
}

func (qemu *Qemu) Patch(q v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	jsonData, err := json.Marshal(q)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	resp, err := qemu.client.dynamicPatch(qemu.resourceId, q.Metadata.Name, jsonData)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	var result v1alpha1.Qemu
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	return result, nil
}

func (qemu *Qemu) UpdateStatus(q v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	jsonData, err := json.Marshal(q)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	resp, err := qemu.client.dynamicUpdateStatus(qemu.resourceId, q.Metadata.Name, jsonData)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	var result v1alpha1.Qemu
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return v1alpha1.Qemu{}, err
	}

	return result, nil
}
