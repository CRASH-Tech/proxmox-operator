/*
Copyright 2024 The CRASH-Tech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxmox

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"k8s.io/utils/strings/slices"
)

const (
	RESOURCE_QEMU    = "qemu"
	RESOURCE_LXC     = "lxc"
	RESOURCE_OPENVZ  = "openvz"
	RESOURCE_STORAGE = "storage"
	RESOURCE_NODE    = "node"
	RESOURCE_SDN     = "sdn"
	RESOURCE_POOL    = "pool"
)

type ClusterApiConfig struct {
	ApiUrl         string `yaml:"api_url"`
	ApiTokenId     string `yaml:"api_token_id"`
	ApiTokenSecret string `yaml:"api_token_secret"`
	Pool           string `yaml:"pool"`
}

type Cluster struct {
	name      string
	apiCOnfig ClusterApiConfig
	resty     *resty.Client
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
	Tags       string  `json:"tags,omitempty"`
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
	Nodes []NodeResp `json:"data"`
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

type PlaceRequest struct {
	Name         string
	CPU          int
	Mem          int64
	AntiAffinity string
}

func (cluster *Cluster) GetReq(apiPath string, data interface{}) ([]byte, error) {
	resp, err := cluster.resty.R().
		SetBody(data).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/%s", cluster.apiCOnfig.ApiUrl, apiPath))
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return resp.Body(), nil
}

func (cluster *Cluster) PostReq(apiPath string, data interface{}) error {
	resp, err := cluster.resty.R().
		SetBody(data).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		Post(fmt.Sprintf("%s/%s", cluster.apiCOnfig.ApiUrl, apiPath))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s %s", resp.StatusCode(), resp.Status(), resp.Body())
	}

	return nil
}

func (cluster *Cluster) PutReq(apiPath string, data interface{}) error {
	resp, err := cluster.resty.R().
		SetBody(data).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		Put(fmt.Sprintf("%s/%s", cluster.apiCOnfig.ApiUrl, apiPath))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s %s", resp.StatusCode(), resp.Status(), resp.Body())
	}

	return nil
}

func (cluster *Cluster) DeleteReq(apiPath string) error {
	resp, err := cluster.resty.R().
		Delete(fmt.Sprintf("%s/%s", cluster.apiCOnfig.ApiUrl, apiPath))
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return nil
}

func (cluster *Cluster) GetNextId() (int, error) {
	log.Debugf("Get next id, cluster: %s", cluster.name)

	apiPath := "/cluster/nextid"

	data, err := cluster.GetReq(apiPath, nil)
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
	data, err := cluster.GetReq(apiPath, reqData)
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

func (cluster *Cluster) GetNodes() ([]NodeResp, error) {
	apiPath := "/nodes"

	data, err := cluster.GetReq(apiPath, nil)
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

func (cluster *Cluster) GetNode(nodeName string) (NodeResp, error) {
	apiPath := "/nodes"

	data, err := cluster.GetReq(apiPath, nil)
	if err != nil {
		return NodeResp{}, err
	}

	nodesData := NodesResp{}
	err = json.Unmarshal(data, &nodesData)
	if err != nil {
		return NodeResp{}, err
	}

	for _, node := range nodesData.Nodes {
		if node.Node == nodeName {
			return node, nil
		}
	}

	return NodeResp{}, err
}

func (cluster *Cluster) GetResourceCount(resourceType string) (int, error) {
	resources, err := cluster.GetResources(resourceType)
	if err != nil {
		return -1, err
	}

	var result int
	for _, r := range resources {
		if r.Type == resourceType {
			result++
		}
	}

	return result, nil
}

func (cluster *Cluster) GetQemuPlacableNode(request PlaceRequest) (string, error) {
	nodes, err := cluster.GetNodes()
	if err != nil {
		return "", err
	}

	var candidateNode string
	var prevCount int
	for _, node := range nodes {
		resources, err := cluster.Node(node.Node).GetResources(RESOURCE_QEMU)
		if err != nil {
			return "", err
		}

		var ignoreNode bool
		for _, resource := range resources {
			tags := strings.Split(resource.Tags, ";")
			if slices.Contains(tags, fmt.Sprintf("anti-affinity.%s", request.AntiAffinity)) {
				ignoreNode = true
			}
		}

		if ignoreNode {
			continue
		}

		qemuCount := len(resources)
		if placable, err := cluster.Node(node.Node).IsQemuPlacable(request.CPU, request.Mem); err == nil && placable {
			if candidateNode == "" || qemuCount < prevCount {
				candidateNode = node.Node
				prevCount = qemuCount
			}
		}
	}

	if candidateNode != "" {
		return candidateNode, nil
	} else {
		return "", fmt.Errorf("cannot find available node")
	}
}

func (cluster *Cluster) FindQemuPlace(name string) (QemuPlace, error) {
	var place QemuPlace

	nodes, err := cluster.GetNodes()
	if err != nil {
		return place, err
	}

	for _, node := range nodes {
		resources, err := cluster.Node(node.Node).GetResources(RESOURCE_QEMU)
		if err != nil {
			return place, err
		}

		for _, resource := range resources {
			if resource.Name == name {
				place.Found = true
				place.Cluster = cluster.name
				place.Node = node.Node
				place.VmId = resource.Vmid

				return place, nil
			}
		}
	}

	return place, fmt.Errorf("cannot find qemu place: %s", name)
}
