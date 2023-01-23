package proxmox

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
	log "github.com/sirupsen/logrus"
)

const (
	ResourceQemu    = "qemu"
	ResourceLXC     = "lxc"
	ResourceOpenVZ  = "openvz"
	ResourceStorage = "storage"
	ResourceNode    = "node"
	ResourceSDN     = "sdn"
	ResourcePool    = "pool"
)

type Cluster struct {
	name      string
	apiConfig common.ApiConfig
}

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

type NodesResp struct {
	Nodes []Node `json:"data"`
}

type NodeResp struct {
	Maxmem         int64   `json:"maxmem"`
	Maxdisk        int64   `json:"maxdisk"`
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	Mem            int64   `json:"mem"`
	Uptime         int     `json:"uptime"`
	SslFingerprint string  `json:"ssl_fingerprint"`
	CPU            float64 `json:"cpu"`
	Level          string  `json:"level"`
	Disk           int64   `json:"disk"`
	Status         string  `json:"status"`
	Maxcpu         int     `json:"maxcpu"`
	Node           string  `json:"node"`
}

func (cluster *Cluster) GetNextId() (int, error) {
	log.Infof("Get next id, cluster: %s", cluster.name)

	apiPath := "/cluster/nextid"

	data, err := common.GetReq(cluster.apiConfig, apiPath, nil)
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

func (cluster *Cluster) GetResources(resourceType string) ([]Resource, error) {
	apiPath := "/cluster/resources"

	reqData := fmt.Sprintf(`{"type":"%s"}`, resourceType)
	data, err := common.GetReq(cluster.apiConfig, apiPath, reqData)
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

func (cluster *Cluster) Node(node string) *Node {
	result := Node{
		name:    node,
		cluster: *cluster,
	}

	return &result
}

func (cluster *Cluster) GetNodes() ([]Node, error) {
	apiPath := "/nodes"

	data, err := common.GetReq(cluster.apiConfig, apiPath, nil)
	if err != nil {
		return nil, err
	}

	nodesData := NodesResp{}
	err = json.Unmarshal(data, &nodesData)
	if err != nil {
		return nil, err
	}

	return nodesData.Nodes, err
}
