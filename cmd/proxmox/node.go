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
