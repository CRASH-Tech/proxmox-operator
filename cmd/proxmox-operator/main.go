package proxmoxoperator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
	v1alpha1 "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api/v1alpha1"
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
		list, err := api.DynamicGetClusterResources(ctx, config.DynamicClient, resourceId)
		if err != nil {
			panic(err)
		}
		for _, item := range list {
			qemu := v1alpha1.QemuImpl{}
			//fmt.Println(item.GetName())
			//fmt.Println(item.Object)
			data, err := item.MarshalJSON()
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(data, &qemu)
			if err != nil {
				panic(err)
			}
			fmt.Println(qemu.Spec)
		}
	}
}
