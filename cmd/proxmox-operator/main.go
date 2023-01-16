package proxmoxoperator

import (
	"context"
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator/api"
)

func Start(config common.Config) {
	ctx := context.Background()
	for {
		fmt.Println(config)
		fmt.Println("lol")
		time.Sleep(time.Second * 1)
		list, err := api.GetResourcesDynamically(config.DynamicClient, ctx, "proxmox.xfix.org", "v1alpha1", "qemu", "sidero-system")
		if err != nil {
			panic(err)
		}
		fmt.Println(list)
	}
}
