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

		// fmt.Println("Update item:")
		// item, err = api.DynamicGetClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu")
		// if err != nil {
		// 	panic(err)
		// }

		// var qemu v1alpha1.QemuImpl
		// err = common.StructCR(item, &qemu)
		// if err != nil {
		// 	panic(err)
		// }
		// qemu.Spec.Accepted = true
		// jsonData, err := json.Marshal(qemu)
		// if err != nil {
		// 	panic(err)
		// }

		// result := unstructured.Unstructured{}
		// err = result.UnmarshalJSON(jsonData)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println("result:", result)
		// item, err = api.DynamicUpdateClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu", result)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(item)

		fmt.Println("Patch item:")
		// data := v1alpha1.QemuImpl{}
		// data.Spec.Accepted = true
		// jsonData, err := json.Marshal(data)
		// if err != nil {
		// 	panic(err)
		// }
		jsonData := []byte(`{"spec":{"accepted":true}}`)
		fmt.Printf("%s\n", jsonData)

		item, err = api.DynamicPatchClusterResource(ctx, config.DynamicClient, resourceId, "example-qemu", jsonData)
		if err != nil {
			panic(err)
		}
		fmt.Println(item)

	}
}
