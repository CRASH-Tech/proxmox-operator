package main

import (
	"context"
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
		log.Error(err)
		return
	}

	for _, qemu := range qemus {
		if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_EMPTY && qemu.Metadata.DeletionTimestamp == "" {
			createNewQemu(kCLient, pClient, qemu)
			return
		}
		if qemu.Metadata.DeletionTimestamp != "" {
			deleteQemu(kCLient, pClient, qemu)
			return
		}
		if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_DEPLOYED {
			syncQemu(kCLient, pClient, qemu)
			return
		}
	}

}

func createNewQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	var place proxmox.QemuPlace
	if qemu.Spec.Cluster == "" {
		if qemu.Spec.Pool == "" {
			log.Error("no cluster or pool are set for: %s", qemu.Metadata.Name)
			return
		}

		var err error
		place, err = pClient.GetQemuPlacableCluster((qemu.Spec.CPU.Cores * qemu.Spec.CPU.Sockets), qemu.Spec.Memory.Size)
		if err != nil {
			log.Error(err)
			return
		}
		qemu.Status.Cluster = place.Cluster
	}

	if qemu.Spec.Node == "" {
		qemu.Status.Node = place.Node
	} else {
		qemu.Status.Node = qemu.Spec.Node
	}

	if qemu.Spec.VmId == 0 {
		qemu.Status.VmId = place.VmId
	} else {
		qemu.Status.VmId = qemu.Spec.VmId
	}

	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		log.Error(err)
		return
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Create(qemuConfig)
	if err != nil {
		log.Error(err)
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		_, err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			log.Error(err)
		}
	}

	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
	_, err = kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		log.Error(err)
	}
}

func deleteQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DELETING
	qemu, err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		log.Error(err)
		return
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Delete(qemu.Status.VmId)
	if err != nil {
		log.Error(err)
		return
	}

	qemu.RemoveFinalizers()
	_, err = kCLient.V1alpha1().Qemu().Patch(qemu)
	if err != nil {
		log.Error(err)
		return
	}
}

func syncQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	qemuStatus, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetStatus(qemu.Status.VmId)
	if err != nil {
		log.Error(err)
		return
	}

	if qemuStatus.Data.Status == proxmox.STATUS_RUNNING {
		qemu.Status.Power = v1alpha1.STATUS_POWER_ON
	} else if qemuStatus.Data.Status == proxmox.STATUS_STOPPED {
		qemu.Status.Power = v1alpha1.STATUS_POWER_OFF
	} else {
		qemu.Status.Power = v1alpha1.STATUS_POWER_UNKNOWN
	}

	_, err = kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		log.Error(err)
		return
	}
}

func buildQemuConfig(client *proxmox.Client, qemu v1alpha1.Qemu) (proxmox.QemuConfig, error) {
	result := make(map[string]interface{})

	result["vmid"] = qemu.Status.VmId
	result["node"] = qemu.Status.Node
	result["name"] = qemu.Metadata.Name
	result["cpu"] = qemu.Spec.CPU.Type
	result["sockets"] = qemu.Spec.CPU.Sockets
	result["cores"] = qemu.Spec.CPU.Cores
	result["memory"] = qemu.Spec.Memory.Size
	result["balloon"] = qemu.Spec.Memory.Balloon

	for _, iface := range qemu.Spec.Network {
		if iface.Mac == "" {
			result[iface.Name] = fmt.Sprintf("model=%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
		} else {
			result[iface.Name] = fmt.Sprintf("model=%s,macaddr=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
		}
	}

	for _, disk := range qemu.Spec.Disk {
		r := regexp.MustCompile(`^[a-z]+(\d+)$`)
		diskNum := r.FindStringSubmatch(disk.Name)
		if len(diskNum) != 2 {
			return nil, fmt.Errorf("cannot extract disk num: %s", disk.Name)
		}
		filename := fmt.Sprintf("vm-%d-disk-%s", qemu.Status.VmId, diskNum[1])
		storageConfig := proxmox.StorageConfig{
			Node:     qemu.Status.Node,
			VmId:     qemu.Status.VmId,
			Filename: filename,
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
		err := client.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).StorageCreate(storageConfig)
		if err != nil {
			return proxmox.QemuConfig{}, err
		}
		result[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, filename, disk.Size)
	}

	for k, v := range qemu.Spec.Options {
		result[k] = v
	}

	return result, nil
}
