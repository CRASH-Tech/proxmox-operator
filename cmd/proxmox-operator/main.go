package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Start(config common.Config) {
	ctx := context.Background()
	resourceId := schema.GroupVersionResource{
		Group:    "proxmox.xfix.org",
		Version:  "v1alpha1",
		Resource: "qemu",
	}

	for {
		//fmt.Println(config)
		fmt.Println("lol")
		time.Sleep(time.Second * 1)

		fmt.Println("Get items:")
		items, err := api.DynamicGetClusterResources(ctx, config.DynamicClient, resourceId)
		if err != nil {
			panic(err)
		}
		fmt.Println(items)

		fmt.Println("Get item:")
		item, err := api.DynamicGetClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu")
		if err != nil {
			panic(err)
		}
		fmt.Println(item)

	}
}
