package common

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Clusters      map[string]common.ApiConfig `yaml:"clusters"`
	DynamicClient *dynamic.DynamicClient
}

func StructCR(obj unstructured.Unstructured, out interface{}) error {
	objJson, err := obj.MarshalJSON()
	if err != nil {
		return err
	}

	err = json.Unmarshal(objJson, &out)
	if err != nil {
		return err
	}

	return nil
}
