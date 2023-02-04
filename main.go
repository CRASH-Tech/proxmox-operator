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
			err = createNewQemu(kCLient, pClient, qemu)
			if err != nil {
				log.Error("cannot create qemu vm %s", qemu.Metadata.Name, err)
			}
			return
		}
		if qemu.Metadata.DeletionTimestamp != "" {
			err = deleteQemu(kCLient, pClient, qemu)
			if err != nil {
				log.Error("cannot delete qemu vm %s", qemu.Metadata.Name, err)
			}
			return
		}

		qemu, err = syncQemuPlaceStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Error("cannot sync qemu deploy status", err)
		}

		qemu, err = syncQemuPowerStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Error("cannot sync qemu power status", err)
		}

		qemu, err = syncQemuNetStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Error("cannot sync qemu network status", err)
		}

		qemu, err = syncQemuDeployStatus(kCLient, pClient, qemu)
		if err != nil {
			log.Error("cannot sync qemu deploy status", err)
		}

	}

}

func createNewQemu(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) error {
	if place, _ := pClient.GetQemuPlaca(qemu.Metadata.Name); place.Cluster != "" {
		log.Info("Qemu already exist, skip creation:", place)
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId
		_, err := kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return fmt.Errorf("cannot get qemu avialable place: %s", err)
		}
		return nil
	}

	if qemu.Spec.Cluster == "" && qemu.Spec.Pool == "" {
		return fmt.Errorf("no cluster or pool are set for: %s", qemu.Metadata.Name)
	}

	var place proxmox.QemuPlace
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

	for _, disk := range qemu.Spec.Disk {
		r := regexp.MustCompile(`^[a-z]+(\d+)$`)
		diskNum := r.FindStringSubmatch(disk.Name)
		if len(diskNum) != 2 {
			return fmt.Errorf("cannot extract disk num: %s", disk.Name)
		}
		filename := fmt.Sprintf("vm-%d-disk-%s", qemu.Status.VmId, diskNum[1])
		storageConfig := proxmox.StorageConfig{
			Node:     qemu.Status.Node,
			VmId:     qemu.Status.VmId,
			Filename: filename,
			Size:     disk.Size,
			Storage:  disk.Storage,
		}
		err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).StorageCreate(storageConfig)
		if err != nil {
			return fmt.Errorf("cannot create qemu disk: %s", err)
		}
		qemuConfig[disk.Name] = fmt.Sprintf("%s:%s,size=%s", disk.Storage, filename, disk.Size)
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Create(qemuConfig)
	if err != nil {
		qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_ERROR
		_, err := kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return fmt.Errorf("cannot create qemu: %s", err)
		}

		return fmt.Errorf("cannot create qemu: %s", err)
	}

	qemu.Status.Deploy = v1alpha1.STATUS_DEPLOY_DEPLOYED
	_, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
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
		log.Info("Waiting qemu stop for deletion: %s", qemu.Metadata.Name)
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

func syncQemuDeployStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	fmt.Println(pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetPendingConfig(qemu.Status.VmId))

	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, err
	}

	designConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		return qemu, err
	}

	var outOfSync bool
	for k, v := range designConfig {
		if k == "node" || k == "vmid" {
			continue
		}
		if fmt.Sprint(currentConfig[k]) != fmt.Sprint(v) {
			log.Infof("Qemu %s is out of sync, %s: %v != %v", qemu.Metadata.Name, k, currentConfig[k], v)
			outOfSync = true
		}
	}

	var syncFail bool
	if outOfSync {
		err = setQemuConfig(kClient, pClient, qemu)
		if err != nil {
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
		return qemu, err
	}

	return qemu, nil
}

func syncQemuPlaceStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if place, _ := pClient.GetQemuPlaca(qemu.Metadata.Name); place.Cluster != "" {
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId

		var err error
		qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
		if err != nil {
			return qemu, err
		}
	}

	return qemu, nil
}

func syncQemuPowerStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	qemuStatus, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetStatus(qemu.Status.VmId)
	if err != nil {
		return qemu, err
	}
	if qemuStatus.Data.Status == proxmox.STATUS_RUNNING {
		qemu.Status.Power = v1alpha1.STATUS_POWER_ON
	} else if qemuStatus.Data.Status == proxmox.STATUS_STOPPED {
		qemu.Status.Power = v1alpha1.STATUS_POWER_OFF
	} else {
		qemu.Status.Power = v1alpha1.STATUS_POWER_UNKNOWN
	}

	qemu, err = kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return qemu, err
	}

	return qemu, nil
}

func syncQemuNetStatus(kClient *kuberentes.Client, pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
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
				//result[iface.Name] = fmt.Sprintf("model=%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
				result[iface.Name] = fmt.Sprintf("%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
			} else {
				//result[iface.Name] = fmt.Sprintf("model=%s,macaddr=%s,bridge=%s,tag=%d", iface.Model, ifaceCurrentMacs[iface.Name], iface.Bridge, iface.Tag)
				result[iface.Name] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, ifaceCurrentMacs[iface.Name], iface.Bridge, iface.Tag)
			}
		} else {
			//result[iface.Name] = fmt.Sprintf("model=%s,macaddr=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
			result[iface.Name] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
		}
	}

	for k, v := range qemu.Spec.Options {
		result[k] = v
	}

	return result, nil
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
