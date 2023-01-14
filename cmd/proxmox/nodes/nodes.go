package nodes

import (
	"encoding/json"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type NodesResp struct {
	Nodes []Node `json:"data"`
}

type Node struct {
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

func Get(apiConfig common.ApiConfig) ([]Node, error) {
	apiPath := "/nodes"

	data, err := common.GetReq(apiConfig, apiPath, nil)
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
