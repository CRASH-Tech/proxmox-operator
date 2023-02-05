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
		log.Fatal(err)
	}
	config = c

	switch config.Log.Format {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	switch config.Log.Level {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	var restConfig *rest.Config
	if path, isSet := os.LookupEnv("KUBECONFIG"); isSet {
		log.Printf("using configuration from '%s'", path)
		restConfig, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
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

func processV1aplha1(kClient *kuberentes.Client, pClient *proxmox.Client) {
	log.Info("Refreshing v1alpha1...")
	qemus, err := kClient.V1alpha1().Qemu().GetAll()
	if err != nil {
		log.Error(err)
		return
	}

	for _, qemu := range qemus {
		switch qemu.Status.Deploy {
		case v1alpha1.STATUS_DEPLOY_EMPTY, v1alpha1.STATUS_DEPLOY_ERROR:
			qemu, err := getQemuPlace(pClient, qemu)
			if err != nil {
				log.Error("cannot get qemu place %s: %s", qemu.Metadata.Name, err)
				qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
				qemu = cleanQemuPlaceStatus(qemu)
				qemu = updateQemuStatus(kClient, qemu)

				continue
			}

			qemu, err = createNewQemu(pClient, qemu)
			if err != nil {
				log.Error("cannot create qemu %s: %s", qemu.Metadata.Name, err)
				if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_ERROR {
					qemu = cleanQemuPlaceStatus(qemu)
				}

				qemu = updateQemuStatus(kClient, qemu)

				continue
			}

			qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_SYNCED
			qemu = updateQemuStatus(kClient, qemu)
		case v1alpha1.STATUS_DEPLOY_SYNCED, v1alpha1.STATUS_DEPLOY_NOT_SYNCED, v1alpha1.STATUS_DEPLOY_PENDING, v1alpha1.STATUS_DEPLOY_UNKNOWN:
			place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
			if err != nil {
				log.Errorf("cannot get qemu place %s %s", qemu.Metadata.Name, err)
				qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_UNKNOWN
				qemu = cleanQemuPlaceStatus(qemu)
				qemu = updateQemuStatus(kClient, qemu)

				continue
			}

			if !place.Found {
				qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_UNKNOWN
				qemu = cleanQemuPlaceStatus(qemu)
				qemu = updateQemuStatus(kClient, qemu)

				continue
			} else {

			}
		default:
			log.Warnf("unknown qemu state: %s %s", qemu.Metadata.Name, qemu.Status.Deploy)
		}
	}

}

func cleanQemuPlaceStatus(qemu v1alpha1.Qemu) v1alpha1.Qemu {
	qemu.Status.Cluster = ""
	qemu.Status.Node = ""
	qemu.Status.VmId = -1

	return qemu
}

func updateQemuStatus(kClient *kuberentes.Client, qemu v1alpha1.Qemu) v1alpha1.Qemu {
	var err error
	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		log.Errorf("cannot update qemu status %s: %s", qemu.Metadata.Name, err)
	}
	return qemu
}

func getQemuPlace(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
	if err != nil {
		return qemu, fmt.Errorf("cannot check is qemu already exist: %s", err)
	}

	if place.Found {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_NOT_SYNCED
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId

		return qemu, fmt.Errorf("qemu already exist, skip creation: %s place: %v", qemu.Metadata.Name, place)
	}

	if qemu.Spec.Cluster == "" && qemu.Spec.Pool == "" {
		return qemu, fmt.Errorf("no cluster or pool are set for: %s", qemu.Metadata.Name)
	}

	if qemu.Spec.Pool != "" {
		var err error
		place, err = pClient.GetQemuPlacableCluster((qemu.Spec.CPU.Cores * qemu.Spec.CPU.Sockets), qemu.Spec.Memory.Size)
		if err != nil {
			return qemu, fmt.Errorf("cannot find autoplace cluster: %s", err)
		}
	}

	if qemu.Spec.Cluster == "" {
		qemu.Status.Cluster = place.Cluster
	} else {
		qemu.Status.Cluster = qemu.Spec.Cluster
	}

	if qemu.Spec.Node == "" {
		node, err := pClient.Cluster(qemu.Status.Cluster).GetQemuPlacableNode((qemu.Spec.CPU.Cores * qemu.Spec.CPU.Sockets), qemu.Spec.Memory.Size)
		if err != nil {
			return qemu, fmt.Errorf("cannot find avialable node: %s", err)
		}
		qemu.Status.Node = node
	} else {
		qemu.Status.Node = qemu.Spec.Node
	}

	if qemu.Spec.VmId == 0 {
		nextId, err := pClient.Cluster(qemu.Status.Cluster).GetNextId()
		if err != nil {
			return qemu, fmt.Errorf("cannot get qemu next id: %s", err)
		}
		qemu.Status.VmId = nextId
	} else {
		qemu.Status.VmId = qemu.Spec.VmId
	}

	return qemu, nil
}

func createNewQemu(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		return qemu, fmt.Errorf("cannot build qemu config: %s", err)
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Create(qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		return qemu, fmt.Errorf("cannot create qemu: %s", err)
	}

	qemuConfig, err = createQemuDisks(pClient, qemu, qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_NOT_SYNCED
		return qemu, fmt.Errorf("cannot create qemu disk: %s", err)
	}

	pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().SetConfig(qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_NOT_SYNCED
		return qemu, fmt.Errorf("cannot set qemu config: %s", err)
	}

	if qemu.Spec.Autostart {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Start(qemu.Status.VmId)
		if err != nil {
			log.Errorf("cannot start qemu: %s %s", qemu.Metadata.Name, err)
			return qemu, nil
		}
	}

	return qemu, nil
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

	ifaceCurrentMacs := make(map[string]string)
	for _, data := range qemu.Status.Net {
		ifaceCurrentMacs[data.Name] = data.Mac
	}

	for _, iface := range qemu.Spec.Network {
		if iface.Mac == "" {
			if ifaceCurrentMacs[iface.Name] == "" {
				result[iface.Name] = fmt.Sprintf("%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
			} else {
				result[iface.Name] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, ifaceCurrentMacs[iface.Name], iface.Bridge, iface.Tag)
			}
		} else {
			result[iface.Name] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
		}
	}

	for k, v := range qemu.Spec.Options {
		result[k] = v
	}

	return result, nil
}

func createQemuDisks(pClient *proxmox.Client, qemu v1alpha1.Qemu, qemuConfig proxmox.QemuConfig) (proxmox.QemuConfig, error) {
	for _, disk := range qemu.Spec.Disk {
		storageConfig, err := buildStorageConfig(qemu)
		if err != nil {
			return qemuConfig, fmt.Errorf("cannot build storage config: %s", err)
		}
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).StorageCreate(storageConfig)
		if err != nil {
			return qemuConfig, fmt.Errorf("cannot create qemu disk: %s", err)
		}
		qemuConfig[disk.Name] = fmt.Sprintf("%s:%s,size=%s", storageConfig.Storage, storageConfig.Filename, storageConfig.Size)
	}

	return qemuConfig, nil
}

func buildStorageConfig(qemu v1alpha1.Qemu) (proxmox.StorageConfig, error) {
	var storageConfig proxmox.StorageConfig

	rDiskNum := regexp.MustCompile(`^[a-z]+(\d+)$`)
	for _, disk := range qemu.Spec.Disk {
		diskNum := rDiskNum.FindStringSubmatch(disk.Name)
		if len(diskNum) != 2 {
			return storageConfig, fmt.Errorf("cannot extract disk num: %s", disk.Name)
		}
		filename := fmt.Sprintf("vm-%d-disk-%s", qemu.Status.VmId, diskNum[1])
		storageConfig = proxmox.StorageConfig{
			Node:     qemu.Status.Node,
			VmId:     qemu.Status.VmId,
			Filename: filename,
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
	}
	return storageConfig, nil
}
