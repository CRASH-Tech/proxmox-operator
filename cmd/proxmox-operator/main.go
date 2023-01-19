package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	papi "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"

	v1alpha1 "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api/v1alpha1"
)

func Start(config common.Config) {
	ctx := context.Background()
	pApi := papi.New(ctx, *config.DynamicClient)

	for {
		//fmt.Println(config)
		fmt.Println("lol")
		time.Sleep(time.Second * 1)

		fmt.Println("Get item:")
		qemu, err := v1alpha1.QemuGet(*pApi, "example-qemu")
		if err != nil {
			panic(err)
		}
		fmt.Println(qemu.Spec)

		fmt.Println("Get items:")
		items, err := v1alpha1.QemuGetAll(*pApi)
		if err != nil {
			panic(err)
		}

		for _, qemu := range items {
			fmt.Println(qemu.Metadata.Name)
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

	}
}
