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

type Node struct {
	name    string
	cluster Cluster
}

func (node *Node) Qemu() *Qemu {
	result := Qemu{
		node: node,
	}

	return &result
}

func (node *Node) GetResources(resourceType string) ([]Resource, error) {
	resources, err := node.cluster.GetResources(resourceType)
	if err != nil {
		return nil, err
	}

	var result []Resource
	for _, resource := range resources {
		if resource.Node == node.name {
			result = append(result, resource)
		}
	}

	return result, nil
}

func (node *Node) GetResourceCount(resourceType string) (int, error) {
	resources, err := node.cluster.GetResources(resourceType)
	if err != nil {
		return -1, err
	}

	var result int
	for _, resource := range resources {
		if resource.Node == node.name && resource.Type == resourceType {
			result++
		}
	}

	return result, nil
}

func (node *Node) IsQemuPlacable(cpu int, mem int64) (bool, error) {
	nodeResources, err := node.cluster.GetNode(node.name)
	if err != nil {
		return false, err
	}

	if (float64(nodeResources.Maxcpu)-nodeResources.CPU) > float64(cpu) && (nodeResources.Maxmem-nodeResources.Mem > mem) {
		return true, nil
	}

	return false, nil
}
