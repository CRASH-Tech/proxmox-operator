package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	kuberentes "github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes/api/v1alpha1"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version = "0.0.1"
	config  common.Config
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

	ctx := context.Background()
	kClient := kuberentes.NewClient(ctx, *config.DynamicClient)
	pClient := proxmox.NewClient(config.Clusters)

	for {
		processV1aplha1(kClient, pClient)

		time.Sleep(5 * time.Second)
	}
}

func readConfig(path string) (common.Config, error) {
	config := common.Config{}
	config.Clusters = make(map[string]proxmox.ClusterApiConfig)

	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return common.Config{}, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return common.Config{}, err
	}

	return config, err
}

func processV1aplha1(kCLient *kuberentes.Client, pClient *proxmox.Client) {
	log.Info("Refreshing v1alpha1...")

	qemus, err := kCLient.V1alpha1().Qemu().GetAll()
	if err != nil {
		panic(err)
	}

	for _, qemu := range qemus {
		fmt.Println(qemu.Status)
		if qemu.Status.Deploy == "" {
			qemu.Status.Deploy = "Deploying"
			qemu.Status.Cluster = "test"
			qemu.RemoveFinalizers()
			//err := kCLient.V1alpha1().Qemu().Patch(qemu)
			err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
			if err != nil {
				panic(err)
			}
		}

	}

}

func buildQemuConfig(client *proxmox.Client, cr v1alpha1.Qemu) (proxmox.QemuConfig, error) {
	result := make(map[string]interface{})

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
		r := regexp.MustCompile(`^[a-z]+(\d+)$`)
		diskNum := r.FindStringSubmatch(disk.Name)
		if len(diskNum) != 2 {
			return nil, errors.New(fmt.Sprintf("cannot extract disk num: %s", disk.Name))
		}
		filename := fmt.Sprintf("vm-%d-disk-%s", cr.Spec.Vmid, diskNum[1])
		fmt.Println(filename)
		storageConfig := proxmox.StorageConfig{
			Node:     cr.Spec.Node,
			VmId:     cr.Spec.Vmid,
			Filename: filename,
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
		err := client.Cluster(cr.Spec.Cluster).Node(cr.Spec.Node).StorageCreate(storageConfig)
		if err != nil {
			panic(err)
		}
		result[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, "vm-222-disk-0", disk.Size)
	}

	for k, v := range cr.Spec.Options {
		result[k] = v
	}

	return result, nil
}
