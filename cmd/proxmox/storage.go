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
	"fmt"

	log "github.com/sirupsen/logrus"
)

type DiskConfig struct {
	Name     string `json:"-"`
	Node     string `json:"node"`
	VmId     int    `json:"vmid"`
	Filename string `json:"filename"`
	Size     string `json:"size"`
	Storage  string `json:"storage"`
}

func (node *Node) DiskCreate(diskConfig DiskConfig) error {
	log.Debugf("Creating disk, cluster: %s, node: %s config: %+v", node.cluster.name, node.name, diskConfig)
	apiPath := fmt.Sprintf("/nodes/%s/storage/%s/content", node.name, diskConfig.Storage)
	err := node.cluster.PostReq(apiPath, diskConfig)
	if err != nil {
		return err
	}

	return nil
}
