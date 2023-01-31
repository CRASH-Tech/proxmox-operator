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
	for _, r := range resources {
		if r.Type == resourceType {
			result++
		}
	}

	return result, nil
}
