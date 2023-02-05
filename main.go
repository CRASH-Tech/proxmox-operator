package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

func processV1aplha1(kCLient *kuberentes.Client, pClient *proxmox.Client) {
	log.Info("Refreshing v1alpha1...")
	qemus, err := kCLient.V1alpha1().Qemu().GetAll()
	if err != nil {
		log.Error(err)
		return
	}

	for _, qemu := range qemus {
		if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_EMPTY &&
			qemu.Metadata.DeletionTimestamp == "" {

			err = createNewQemu(kCLient, pClient, qemu)
			if err != nil {
				log.Errorf("cannot create qemu: %s", err)
				continue
			}
			continue
		}
		if qemu.Metadata.DeletionTimestamp != "" {
			err = deleteQemu(kCLient, pClient, qemu)
			if err != nil {
				log.Errorf("cannot delete qemu: %s", err)
				continue
			}
			continue
		}
		log.Warnf("%v///////////////////////////////////////////////////////////////////////////////////////////////////", qemu)

		qemu, err = syncQemuPlaceStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Error("cannot sync qemu deploy status: %s", err)
		}

		qemu, err = syncQemuPowerStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Errorf("cannot sync qemu power status: %s", err)
		}

		qemu, err = syncQemuNetStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Errorf("cannot sync qemu network status: %s", err)
		}

		qemu, err = syncQemuDisksStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Errorf("cannot sync qemu disks status: %s", err)
		}

		qemu, err = syncQemuDeployStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Errorf("cannot sync qemu deploy status: %s", err)
		}

	}

}

func createNewQemu(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) error {
	place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
	if err != nil {
		return fmt.Errorf("cannot check is qemu already exist: %s", err)
	}
	log.Error(qemu.Metadata.Name, place) //////////////////////////////////////
	if place.Found {
		log.Warnf("Qemu already exist, skip creation: %s place: %v", qemu.Metadata.Name, place)
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return fmt.Errorf("cannot get qemu avialable place: %v", err)
		}
		return nil
	}

	if qemu.Spec.Cluster == "" && qemu.Spec.Pool == "" {
		return fmt.Errorf("no cluster or pool are set for: %s", qemu.Metadata.Name)
	}

	if qemu.Spec.Pool != "" {
		var err error
		place, err = pClient.GetQemuPlacableCluster((qemu.Spec.CPU.Cores * qemu.Spec.CPU.Sockets), qemu.Spec.Memory.Size)
		if err != nil {
			return fmt.Errorf("cannot find autoplace cluster: %s", err)
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
			return fmt.Errorf("cannot find avialable node: %s", err)
		}
		qemu.Status.Node = node
	} else {
		qemu.Status.Node = qemu.Spec.Node
	}

	if qemu.Spec.VmId == 0 {
		nextId, err := pClient.Cluster(qemu.Status.Cluster).GetNextId()
		if err != nil {
			return fmt.Errorf("cannot get qemu next id: %s", err)
		}
		qemu.Status.VmId = nextId
	} else {
		qemu.Status.VmId = qemu.Spec.VmId
	}

	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		return fmt.Errorf("cannot build qemu config: %s", err)
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Create(qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		qemu.Status.Cluster = ""
		qemu.Status.Node = ""
		qemu.Status.VmId = 0
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return fmt.Errorf("cannot create qemu: %s", err)
		}

		return fmt.Errorf("cannot create qemu: %s", err)
	}

	qemuConfig, err = createQemuDisks(kClient, pClient, qemu, qemuConfig)
	if err != nil {
		return fmt.Errorf("cannot create qemu disk: %s", err)
	}

	pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().SetConfig(qemuConfig)
	if err != nil {
		return fmt.Errorf("cannot set qemu config: %s", err)
	}

	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return fmt.Errorf("cannot update new qemu status: %s", err)
	}

	if qemu.Spec.Autostart {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Start(qemu.Status.VmId)
		if err != nil {
			return fmt.Errorf("cannot start qemu: %s", err)
		}
	}

	return nil
}

func deleteQemu(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) error {
	if qemu.Status.Deploy == v1alpha1.STATUS_DEPLOY_EMPTY {
		qemu.RemoveFinalizers()
		_, err := kClient.V1alpha1().Qemu().Patch(qemu)
		if err != nil {
			return fmt.Errorf("cannot remove qemu finalizer: %s", err)
		}
		return nil
	}

	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DELETING
	qemu, err := kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return fmt.Errorf("cannot update qemu status: %s", err)
	}

	if !isQemuPlaced(qemu) {
		return nil
	}

	if qemu.Spec.Autostop {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Stop(qemu.Status.VmId)
		if err != nil {
			return fmt.Errorf("cannot stop: %s", err)
		}
	}

	qemu, err = syncQemuPowerStatus(kClient, pClient, qemu)
	if err != nil {
		return fmt.Errorf("cannot get qemu power status: %s", err)
	}
	if qemu.Status.Power == v1alpha1.STATUS_POWER_ON {
		log.Warnf("Waiting qemu stop for deletion: %s", qemu.Metadata.Name)
	} else {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Delete(qemu.Status.VmId)
		if err != nil {
			return fmt.Errorf("cannot delete qemu: %s", err)
		}

		qemu.RemoveFinalizers()
		_, err = kClient.V1alpha1().Qemu().Patch(qemu)
		if err != nil {
			return fmt.Errorf("cannot remove qemu finalizer: %s", err)
		}
	}

	return nil
}

func syncQemuPlaceStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
	if err != nil {
		return qemu, fmt.Errorf("cannot get qemu place: %s", err)
	}

	if place.Found {
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId
	} else {
		qemu.Status.Cluster = ""
		qemu.Status.Node = ""
		qemu.Status.VmId = 0
	}

	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return qemu, err
	}

	return qemu, nil
}

func syncQemuDeployStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if !isQemuPlaced(qemu) {
		return qemu, nil
	}

	pendingConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetPendingConfig(qemu.Status.VmId)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return qemu, err
		}
		return qemu, fmt.Errorf("cannot get pending config: %s", err)
	}

	var isPending bool
	for _, v := range pendingConfig {
		if v.Pending != nil {
			log.Warnf("Qemu %s is in pending state, %s: %v != %v", qemu.Metadata.Name, v.Key, v.Value, v.Pending)
			isPending = true
		}
	}

	if isPending {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_PENDING
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
			qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
			if err != nil {
				return qemu, err
			}
			return qemu, err
		}

		return qemu, nil
	}

	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return qemu, err
		}
		return qemu, err
	}

	designConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return qemu, err
		}
		return qemu, err
	}

	var outOfSync bool
	for k, v := range designConfig {
		if k == "node" || k == "vmid" {
			continue
		}
		if fmt.Sprint(currentConfig[k]) != fmt.Sprint(v) {
			log.Warnf("Qemu %s is out of sync, %s: %v != %v", qemu.Metadata.Name, k, currentConfig[k], v)
			outOfSync = true
		}
	}

	var syncFail bool
	if outOfSync {
		err = setQemuConfig(kClient, pClient, qemu)
		if err != nil {
			qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
			qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
			if err != nil {
				return qemu, err
			}
			syncFail = true
		}
	}

	if syncFail {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_NOT_SYNCED
	} else {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
	}

	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return qemu, err
		}
		return qemu, err
	}

	return qemu, nil
}

func syncQemuPowerStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if !isQemuPlaced(qemu) {
		return qemu, nil
	}

	qemuStatus, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetStatus(qemu.Status.VmId)
	if err != nil {
		qemu.Status.Power = v1alpha1.STATUS_POWER_UNKNOWN
	}

	if qemuStatus.Data.Status == proxmox.STATUS_RUNNING {
		qemu.Status.Power = v1alpha1.STATUS_POWER_ON
	} else if qemuStatus.Data.Status == proxmox.STATUS_STOPPED {
		qemu.Status.Power = v1alpha1.STATUS_POWER_OFF
	}

	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return qemu, err
	}

	return qemu, nil
}

func syncQemuNetStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if !isQemuPlaced(qemu) {
		return qemu, nil
	}

	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, err
	}

	var ifacesConfig []v1alpha1.QemuStatusNetwork
	rExp := "(.{2}:.{2}:.{2}:.{2}:.{2}:.{2})"
	r := regexp.MustCompile(rExp)
	for _, iface := range qemu.Spec.Network {
		var ifaceConfig v1alpha1.QemuStatusNetwork
		ifaceConfig.Name = iface.Name
		macData := r.FindStringSubmatch(fmt.Sprintf("%s", currentConfig[iface.Name]))
		if len(macData) > 0 {
			ifaceConfig.Mac = macData[0]
		}

		ifacesConfig = append(ifacesConfig, ifaceConfig)
	}
	qemu.Status.Net = ifacesConfig

	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return qemu, err
	}

	return qemu, nil
}

func syncQemuDisksStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if !isQemuPlaced(qemu) {
		return qemu, nil
	}

	qemuConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, err
	}

	rDiskSize := regexp.MustCompile(`^.+size=(.+),?$`)
	for _, disk := range qemu.Spec.Disk {
		var designStorageConfig proxmox.StorageConfig
		designStorageConfig, err = buildStorageConfig(pClient, qemu)
		if err != nil {
			return qemu, fmt.Errorf("cannot build storage config: %s", err)
		}

		for k, v := range qemuConfig {
			if strings.Contains(fmt.Sprint(v), designStorageConfig.Filename) {
				currentSize := rDiskSize.FindStringSubmatch(fmt.Sprint(v))
				if len(currentSize) != 2 {
					return qemu, fmt.Errorf("cannot extract disk num: %s", disk.Name)
				}
				if designStorageConfig.Size != currentSize[1] {
					err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Resize(qemu.Status.VmId, k, designStorageConfig.Size)
					if err != nil {
						return qemu, fmt.Errorf("cannot resize qemu disk: %s", err)
					}
				}
			}
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

func buildStorageConfig(client *proxmox.Client, qemu v1alpha1.Qemu) (proxmox.StorageConfig, error) {
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

func setQemuConfig(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) error {
	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		return err
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().SetConfig(qemuConfig)
	if err != nil {
		return err
	}

	return nil
}

func createQemuDisks(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu, qemuConfig proxmox.QemuConfig) (proxmox.QemuConfig, error) {
	for _, disk := range qemu.Spec.Disk {
		storageConfig, err := buildStorageConfig(pClient, qemu)
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

func isQemuPlaced(qemu v1alpha1.Qemu) bool {
	if qemu.Status.Cluster != "" && qemu.Status.Node != "" && qemu.Status.VmId != 0 {
		return true
	}

	return false
}
