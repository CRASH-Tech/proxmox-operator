package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	. "github.com/CRASH-Tech/proxmox-operator/cmd/common"
	kuberentes "github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version = "0.0.1"
	config  Config
)

func init() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.yaml", "config file path. Default: config.yaml")
	c, err := readConfig(configPath)
	if err != nil {
		panic(err)
	}
	config = c

	var restConfig *rest.Config
	if path, isSet := os.LookupEnv("KUBECONFIG"); isSet {
		log.Printf("using configuration from '%s'", path)
		restConfig, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	}
	config.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
}

func main() {
	log.Infof("Starting proxmox-operator %s\n", version)
	Start()
}

func readConfig(path string) (Config, error) {
	config := Config{}
	config.Clusters = make(map[string]proxmox.ClusterApiConfig)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return Config{}, err
	}

	return config, err
}

func Start() {
	ctx := context.Background()
	kClient := kuberentes.NewClient(ctx, *config.DynamicClient)
	// pClient := proxmox.NewClient(config.Clusters)
	// res, err := pClient.Cluster("crash-lab").GetResources(proxmox.ResourceNode)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(res)
	//client.Cluster("crash-lab").Node("crash-lab").
	//fmt.Println(client)

	//for {
	//fmt.Println(config)
	// fmt.Println("lol")
	// time.Sleep(time.Second * 1)

	fmt.Println("Get item:")
	lol, err := kClient.V1alpha1().Qemu().Get("example-qemu")
	if err != nil {
		panic(err)
	}
	fmt.Println(lol)
	// cr, err := v1alpha1.QemuGet(*kClient, "example-qemu")
	// if err != nil {
	// 	panic(err)
	// }
	//fmt.Println(cr.Spec)

	// fmt.Println("Get items:")
	// crs, err := v1alpha1.QemuGetAll(*kClient)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, cr := range crs {
	// 	fmt.Println(cr.Metadata.Name)
	// 	qemu := buildQemuConfig(pClient, cr)
	// 	err := pClient.Cluster(cr.Spec.Cluster).Node(cr.Spec.Node).Qemu().Create(qemu)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

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

// func buildQemuConfig(client *proxmox.Client, cr v1alpha1.Qemu) (result proxmox.QemuConfig) {
// 	result = make(map[string]interface{})

// 	result["vmid"] = cr.Spec.Vmid
// 	result["node"] = cr.Spec.Node
// 	result["name"] = cr.Metadata.Name
// 	result["cpu"] = cr.Spec.CPU.Type
// 	result["sockets"] = cr.Spec.CPU.Sockets
// 	result["cores"] = cr.Spec.CPU.Cores
// 	result["memory"] = cr.Spec.Memory.Size
// 	result["balloon"] = cr.Spec.Memory.Balloon

// 	for _, iface := range cr.Spec.Network {
// 		if iface.Mac == "" {
// 			result[iface.Name] = fmt.Sprintf("model=%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
// 		} else {
// 			result[iface.Name] = fmt.Sprintf("model=%s,macaddr=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
// 		}
// 	}

// 	for _, disk := range cr.Spec.Disk {
// 		storageConfig := proxmox.StorageConfig{
// 			Node:     cr.Spec.Node,
// 			VmId:     cr.Spec.Vmid,
// 			Filename: "vm-222-disk-0",
// 			Size:     disk.Size,
// 			Storage:  disk.Storage,
// 		}
// 		err := client.Cluster(cr.Spec.Cluster).Node(cr.Spec.Node).StorageCreate(storageConfig)
// 		if err != nil {
// 			panic(err)
// 		}
// 		result[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, "vm-222-disk-0", disk.Size)
// 	}

// 	for k, v := range cr.Spec.Options {
// 		result[k] = v
// 	}

// 	return
// }
