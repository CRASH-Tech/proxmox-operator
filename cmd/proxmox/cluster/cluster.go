package cluster

import (
	"encoding/json"
	"strconv"

	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/common"
)

type NextIdResp struct {
	NextId string `json:"data"`
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
