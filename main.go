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
		panic(err)
	}

	for _, qemu := range qemus {
		if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_EMPTY && qemu.Metadata.DeletionTimestamp == "" {
			go createNewQemu(kCLient, pClient, qemu)
		}
		if qemu.Metadata.DeletionTimestamp != "" && qemu.Status.Deploy != v1alpha1.STATUS_DEPLOY_DELETING {
			deleteQemu(kCLient, pClient, qemu)

		}
		if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_DEPLOYED {
			syncQemu(kCLient, pClient, qemu)
		}
	}

}

func createNewQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	// qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_PROCESSING
	// qemu, err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	// if err != nil {
	// 	panic(err)
	// }

	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		panic(err)
	}

	err = pClient.Cluster(qemu.Spec.Cluster).Node(qemu.Spec.Node).Qemu().Create(qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		_, err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			panic(err)
		}
		panic(err)
	}

	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
	_, err = kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		panic(err)
	}
}

func deleteQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DELETING
	qemu, err := kCLient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		panic(err)
	}

	pClient.Cluster(qemu.Spec.Cluster).Node(qemu.Spec.Node).Qemu().Delete(qemu.Spec.Vmid)
	if err != nil {
		panic(err)
	}

	qemu.RemoveFinalizers()
	_, err = kCLient.V1alpha1().Qemu().Patch(qemu)
	if err != nil {
		panic(err)
	}
}

func syncQemu(kCLient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) {
	qemuStatus, err := pClient.Cluster(qemu.Spec.Cluster).Node(qemu.Spec.Node).Qemu().GetStatus(qemu.Spec.Vmid)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	fmt.Println(qemuStatus)

}

func buildQemuConfig(client *proxmox.Client, qemu v1alpha1.Qemu) (proxmox.QemuConfig, error) {
	result := make(map[string]interface{})

	if qemu.Spec.Cluster == "" {
		if qemu.Spec.Pool == "" {
			return proxmox.QemuConfig{}, fmt.Errorf("no cluster or pool are set for: %s", qemu.Metadata.Name)
		}
		var err error
		qemu, err = findAvialableCluster(client, qemu)
		if err != nil {
			panic(err)
		}
	}

	if qemu.Spec.Node == "" {

	}

	if qemu.Spec.Vmid == 0 {

	}

	result["vmid"] = qemu.Spec.Vmid
	result["node"] = qemu.Spec.Node
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
		filename := fmt.Sprintf("vm-%d-disk-%s", qemu.Spec.Vmid, diskNum[1])
		storageConfig := proxmox.StorageConfig{
			Node:     qemu.Spec.Node,
			VmId:     qemu.Spec.Vmid,
			Filename: filename,
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
		err := client.Cluster(qemu.Spec.Cluster).Node(qemu.Spec.Node).StorageCreate(storageConfig)
		if err != nil {
			panic(err)
		}
		result[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, filename, disk.Size)
	}

	for k, v := range qemu.Spec.Options {
		result[k] = v
	}

	return result, nil
}

type PlaceCandidate struct {
	Cluster   string
	Node      string
	QemuCount int
}

func findAvialableCluster(client *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	//var candidate PlaceCandidate

	for cluster, clusterData := range client.Clusters {
		var c PlaceCandidate
		c.Cluster = cluster

		if clusterData.Pool == qemu.Spec.Pool {
			nodes, err := client.Cluster(cluster).GetResources(proxmox.RESOURCE_NODE)
			if err != nil {
				return qemu, err
			}

			for _, node := range nodes {
				fmt.Println("node")
				fmt.Println(client.Cluster(cluster).Node(node.Node).GetResourceCount(proxmox.RESOURCE_QEMU))
				fmt.Println("cluster")
				fmt.Println(client.Cluster(cluster).GetResourceCount(proxmox.RESOURCE_QEMU))
			}
		}
	}
	return qemu, nil

	// fmt.Println("lol")
	// resources, err := client.GetPoolResources("prod")
	// if err != nil {
	// 	panic(err)
	// }

	// for _, resource := range resources {
	// 	fmt.Println(resource)
	// }
	//return v1alpha1.Qemu{}, nil
}
