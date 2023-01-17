package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	papi "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"

	v1alpha1 "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Start(config common.Config) {
	ctx := context.Background()
	pApi := papi.New(ctx, *config.DynamicClient)
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
		qemu, err := v1alpha1.Get(*pApi, resourceId, "example-qemu")
		if err != nil {
			panic(err)
		}
		fmt.Println(qemu.Metadata.Name)

		fmt.Println("Get items:")
		items, err := v1alpha1.GetAll(*pApi, resourceId)
		if err != nil {
			panic(err)
		}

		for _, qemu := range items {
			fmt.Println(qemu.Metadata.Name)
		}

		fmt.Println("Patch item:")
		qemu, err = v1alpha1.Get(*pApi, resourceId, "example-qemu")
		if err != nil {
			panic(err)
		}
		qemu.Spec.Accepted = true
		err = v1alpha1.Patch(*pApi, resourceId, qemu)
		if err != nil {
			panic(err)
		}

	}
}
