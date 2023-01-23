package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	proxmox "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	papi "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/qemu"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox/nodes/storage"

	v1alpha1 "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api/v1alpha1"
)

func Start(config common.Config) {
	ctx := context.Background()
	pApi := papi.New(ctx, *config.DynamicClient)
	client := proxmox.NewClient(config.Clusters)

	//for {
	//fmt.Println(config)
	fmt.Println("lol")
	time.Sleep(time.Second * 1)

	fmt.Println("Get item:")
	cr, err := v1alpha1.QemuGet(*pApi, "example-qemu")
	if err != nil {
		panic(err)
	}
	fmt.Println(cr.Spec)

	fmt.Println("Get items:")
	crs, err := v1alpha1.QemuGetAll(*pApi)
	if err != nil {
		panic(err)
	}

	for _, cr := range crs {
		fmt.Println(cr.Metadata.Name)
		qemu := buildQemuConfig(client, cr)
		err := client.QemuCreate("crash-lab", qemu)
		if err != nil {
			panic(err)
		}
	}

	// fmt.Println("Patch item:")
	// qemu, err = v1alpha1.QemuGet(*pApi, "example-qemu")
	// if err != nil {
	// 	panic(err)
	// }
	// qemu.Spec.Accepted = false
	// err = v1alpha1.QemuPatch(*pApi, qemu)
	// if err != nil {
	// 	panic(err)
	// }

	//}
}

func buildQemuConfig(client *proxmox.Client, cr v1alpha1.Qemu) (result qemu.QemuConfig) {
	result = make(map[string]interface{})

	result["vmid"] = cr.Spec.Vmid
	result["node"] = cr.Spec.Node
	result["name"] = cr.Metadata.Name
	result["cpu"] = cr.Spec.CPU.Type
	result["sockets"] = cr.Spec.CPU.Sockets
	result["cores"] = cr.Spec.CPU.Cores
	result["memory"] = cr.Spec.Memory.Size
	result["balloon"] = cr.Spec.Memory.Balloon

	for _, iface := range cr.Spec.Network {
		if iface.Mac == "" {
			result[iface.Name] = fmt.Sprintf("model=%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
		} else {
			result[iface.Name] = fmt.Sprintf("model=%s,macaddr=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
		}
	}

	for _, disk := range cr.Spec.Disk {
		storageConfig := storage.StorageConfig{
			Node:     cr.Spec.Node,
			VmId:     cr.Spec.Vmid,
			Filename: "vm-222-disk-0",
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
		err := client.StorageCreate(cr.Spec.Cluster, storageConfig)
		if err != nil {
			panic(err)
		}
		result[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, "vm-222-disk-0", disk.Size)
	}

	for k, v := range cr.Spec.Options {
		result[k] = v
	}

	return
}
