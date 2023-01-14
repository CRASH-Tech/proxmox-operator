package cluster

import (
	"encoding/json"
	"strconv"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type NextIdResp struct {
	NextId string `json:"data"`
}

type ResourcesResp struct {
	Data []Resource `json:"data"`
}

type Resource struct {
	Maxdisk    int64   `json:"maxdisk"`
	Netout     int64   `json:"netout,omitempty"`
	ID         string  `json:"id"`
	Vmid       int     `json:"vmid,omitempty"`
	Type       string  `json:"type"`
	Mem        int64   `json:"mem,omitempty"`
	Diskread   int64   `json:"diskread,omitempty"`
	Maxmem     int64   `json:"maxmem,omitempty"`
	Template   int     `json:"template,omitempty"`
	Status     string  `json:"status"`
	Netin      int64   `json:"netin,omitempty"`
	Maxcpu     int     `json:"maxcpu,omitempty"`
	Node       string  `json:"node"`
	Uptime     int     `json:"uptime,omitempty"`
	Diskwrite  int64   `json:"diskwrite,omitempty"`
	Name       string  `json:"name,omitempty"`
	CPU        float64 `json:"cpu,omitempty"`
	Disk       int     `json:"disk"`
	Level      string  `json:"level,omitempty"`
	Shared     int     `json:"shared,omitempty"`
	Content    string  `json:"content,omitempty"`
	Storage    string  `json:"storage,omitempty"`
	Plugintype string  `json:"plugintype,omitempty"`
}

func GetNextId(apiConfig common.ApiConfig) (int, error) {
	apiPath := "/cluster/nextid"

	data, err := common.GetReq(apiConfig, apiPath, nil)
	if err != nil {
		return -1, err
	}

	nextIdResp := NextIdResp{}
	err = json.Unmarshal(data, &nextIdResp)
	if err != nil {
		return -1, err
	}

	nextId, err := strconv.Atoi(nextIdResp.NextId)
	if err != nil {
		return -1, err
	}

	return nextId, err
}

func GetResources(apiConfig common.ApiConfig) ([]Resource, error) {
	apiPath := "/cluster/resources"

	data, err := common.GetReq(apiConfig, apiPath, nil)
	if err != nil {
		return nil, err
	}

	resourcesResp := ResourcesResp{}
	err = json.Unmarshal(data, &resourcesResp)
	if err != nil {
		return nil, err
	}

	return resourcesResp.Data, err
}
