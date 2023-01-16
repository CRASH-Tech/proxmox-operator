package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"

	v1alpha1 "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Start(config common.Config) {
	ctx := context.Background()
	api := api.New(ctx, *config.DynamicClient)
	resourceId := schema.GroupVersionResource{
		Group:    "proxmox.xfix.org",
		Version:  "v1alpha1",
		Resource: "qemu",
	}

	for {
		//fmt.Println(config)
		fmt.Println("lol")
		time.Sleep(time.Second * 1)

		fmt.Println("Get item:")
		qemu, err := v1alpha1.Get(*api, resourceId, "example-qemu")
		if err != nil {
			panic(err)
		}
		fmt.Println(qemu.Metadata.Name)

		fmt.Println("Get items:")
		items, err := v1alpha1.GetAll(*api, resourceId)
		if err != nil {
			panic(err)
		}

		for _, qemu := range items {
			fmt.Println(qemu.Metadata.Name)
		}

		// fmt.Println("Update(patch) item:")
		// item, err = api.DynamicGetClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu")
		// if err != nil {
		// 	panic(err)
		// }

		// var qemu v1alpha1.Qemu
		// err = common.StructCR(item, &qemu)
		// if err != nil {
		// 	panic(err)
		// }
		// qemu.Spec.Cluster = "lol2"
		// qemu.Spec.Accepted = true
		// qemu.Spec.Config.Agent = false
		// jsonData, err := json.Marshal(qemu)
		// if err != nil {
		// 	panic(err)
		// }

		// fmt.Printf("result: %s", jsonData)
		// item, err = api.DynamicPatchClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu", jsonData)
		// if err != nil {
		// 	panic(err)
		// }
		//fmt.Println(item)

		// fmt.Println("Patch item:")
		// jsonData := []byte(`{"spec":{"accepted":true}}`)
		// fmt.Printf("%s\n", jsonData)

		// item, err = api.DynamicPatchClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu", jsonData)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(item)

	}
}
