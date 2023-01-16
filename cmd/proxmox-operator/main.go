package proxmoxoperator

import (
	"fmt"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
)

func Start(config common.Config) {
	for {
		fmt.Println(config)
		fmt.Println("lol")
		time.Sleep(time.Second * 1)
	}
}
