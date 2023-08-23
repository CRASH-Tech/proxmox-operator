package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/CRASH-Tech/proxmox-operator/cmd/common"
	kubernetes "github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes"
	"github.com/CRASH-Tech/proxmox-operator/cmd/kubernetes/api/v1alpha1"
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	version = "0.1.2"
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
		log.Printf("Using configuration from '%s'", path)
		restConfig, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Using in-cluster configuration")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	config.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
}

func main() {
	log.Infof("Starting proxmox-operator %s", version)

	ctx := context.Background()
	kClient := kubernetes.NewClient(ctx, *config.DynamicClient)
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

func processV1aplha1(kClient *kubernetes.Client, pClient *proxmox.Client) {
	log.Info("Refreshing v1alpha1...")
	qemus, err := kClient.V1alpha1().Qemu().GetAll()
	if err != nil {
		log.Error(err)
		return
	}

	for _, qemu := range qemus {
		switch qemu.Status.Status {
		case v1alpha1.STATUS_QEMU_DELETING:
			qemu, err := deleteQemu(pClient, qemu)
			if err != nil {
				log.Errorf("cannot delete qemu %s: %s", qemu.Metadata.Name, err)

				continue
			}

			qemu.RemoveFinalizers()
			_, err = kClient.V1alpha1().Qemu().Patch(qemu)
			if err != nil {
				log.Errorf("cannot patch qemu cr %s: %s", qemu.Metadata.Name, err)

				continue
			}

			continue
		case v1alpha1.STATUS_QEMU_EMPTY:
			if qemu.Status.Status == v1alpha1.STATUS_QEMU_EMPTY && qemu.Metadata.DeletionTimestamp != "" {
				qemu.RemoveFinalizers()
				_, err = kClient.V1alpha1().Qemu().Patch(qemu)
				if err != nil {
					log.Errorf("cannot patch qemu cr %s: %s", qemu.Metadata.Name, err)

					continue
				}

				continue
			}

			if qemu.Spec.Clone != "" {
				qemu.Status.Status = v1alpha1.STATUS_QEMU_CLONING
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			}

			qemu, err := getQemuPlace(pClient, qemu)
			if err != nil {
				log.Errorf("cannot get qemu place %s: %s", qemu.Metadata.Name, err)

				continue
			}

			if qemu.Status.Status == v1alpha1.STATUS_QEMU_OUT_OF_SYNC {
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			}

			qemu, err = createNewQemu(pClient, qemu)
			if err != nil {
				log.Errorf("cannot create qemu %s: %s", qemu.Metadata.Name, err)
				if qemu.Status.Status == v1alpha1.STATUS_QEMU_EMPTY {
					qemu = cleanQemuPlaceStatus(qemu)
				}

				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			}

			qemu.Status.Status = v1alpha1.STATUS_QEMU_SYNCED
			qemu, err = updateQemuStatus(kClient, qemu)
			if err != nil {
				return
			}

			// Need by proxmox api delay
			time.Sleep(time.Second * 10)

			continue
		case v1alpha1.STATUS_QEMU_CLONING:
			if qemu.Spec.Cluster == "" {
				log.Error("no cluster set for clone operation")

				continue
			}

			templatePlace, err := pClient.Cluster(qemu.Spec.Cluster).FindQemuPlace(qemu.Spec.Clone)
			if err != nil {
				log.Error(err)

				continue
			}

			qemu, err = getQemuPlace(pClient, qemu)
			if err != nil {
				log.Error(err)

				continue
			}

			targetPlace := proxmox.QemuPlace{
				Found:   true,
				Cluster: qemu.Status.Cluster,
				Node:    qemu.Status.Node,
				VmId:    qemu.Status.VmId,
			}

			err = pClient.Cluster(templatePlace.Cluster).Node(templatePlace.Node).Qemu().Clone(qemu.Metadata.Name, templatePlace, targetPlace)
			if err != nil {
				log.Error(err)
				qemu.Status.Status = v1alpha1.STATUS_QEMU_EMPTY
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}
				continue
			}

			qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
			qemu, err = updateQemuStatus(kClient, qemu)
			if err != nil {
				return
			}

			qemu, err = setQemuConfig(pClient, qemu)
			if err != nil {
				log.Errorf("cannot set qemu config %s: %s", qemu.Metadata.Name, err)
				continue
			}

			qemu.Status.Status = v1alpha1.STATUS_QEMU_SYNCED
			qemu, err = updateQemuStatus(kClient, qemu)
			if err != nil {
				return
			}

			// Need by proxmox api delay
			time.Sleep(time.Second * 10)

			continue
		case v1alpha1.STATUS_QEMU_SYNCED,
			v1alpha1.STATUS_QEMU_OUT_OF_SYNC,
			v1alpha1.STATUS_QEMU_PENDING,
			v1alpha1.STATUS_QEMU_UNKNOWN:
			place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
			if err != nil {
				log.Errorf("cannot get qemu place %s: %s", qemu.Metadata.Name, err)
				qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
				qemu = cleanQemuPlaceStatus(qemu)
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			}

			if !place.Found {
				qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
				qemu = cleanQemuPlaceStatus(qemu)
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			} else {
				qemu = updateQemuPlaceStatus(place, qemu)
				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				qemu, err = getQemuPowerStatus(pClient, qemu)
				if err != nil {
					log.Errorf("cannot get qemu power status %s: %s", qemu.Metadata.Name, err)
					qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
					qemu, err = updateQemuStatus(kClient, qemu)
					if err != nil {
						return
					}

					continue
				}

				qemu, err = getQemuNetStatus(pClient, qemu)
				if err != nil {
					log.Errorf("cannot get qemu network status %s: %s", qemu.Metadata.Name, err)
					qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
					qemu, err = updateQemuStatus(kClient, qemu)
					if err != nil {
						return
					}

					continue
				}

				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}
			}

			switch qemu.Status.Status {
			case v1alpha1.STATUS_QEMU_OUT_OF_SYNC, v1alpha1.STATUS_QEMU_PENDING:
				qemu, err := setQemuConfig(pClient, qemu)
				if err != nil {
					log.Errorf("cannot set qemu config %s: %s", qemu.Metadata.Name, err)
				}

				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			default:
				qemu, err = checkQemuSyncStatus(pClient, qemu)
				if err != nil {
					log.Errorf("cannot get qemu sync status %s: %s", qemu.Metadata.Name, err)
					qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
					qemu, err = updateQemuStatus(kClient, qemu)
					if err != nil {
						return
					}

					continue
				}

				qemu, err = updateQemuStatus(kClient, qemu)
				if err != nil {
					return
				}

				continue
			}
		default:
			log.Warnf("unknown qemu state: %s %s", qemu.Metadata.Name, qemu.Status.Status)

			continue
		}
	}

}

func updateQemuPlaceStatus(place proxmox.QemuPlace, qemu v1alpha1.Qemu) v1alpha1.Qemu {
	qemu.Status.Cluster = place.Cluster
	qemu.Status.Node = place.Node
	qemu.Status.VmId = place.VmId

	return qemu
}

func cleanQemuPlaceStatus(qemu v1alpha1.Qemu) v1alpha1.Qemu {
	qemu.Status.Cluster = ""
	qemu.Status.Node = ""
	qemu.Status.VmId = 0

	return qemu
}

func updateQemuStatus(kClient *kubernetes.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	name := qemu.Metadata.Name
	qemu, err := kClient.V1alpha1().Qemu().UpdateStatus(qemu)
	if err != nil {
		return qemu, fmt.Errorf("cannot update qemu status %s: %s", name, err)
	}
	return qemu, nil
}

func getQemuPlace(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	placeRequest := buildPlaceRequest(qemu)
	place, err := pClient.GetQemuPlace(qemu.Metadata.Name)
	if err != nil {
		return qemu, fmt.Errorf("cannot check is qemu already exist: %s", err)
	}

	if place.Found {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
		qemu.Status.Cluster = place.Cluster
		qemu.Status.Node = place.Node
		qemu.Status.VmId = place.VmId

		return qemu, nil
	}

	if qemu.Spec.Cluster == "" && qemu.Spec.Pool == "" {
		return qemu, fmt.Errorf("no cluster or pool are set for: %s", qemu.Metadata.Name)
	}

	if qemu.Spec.Pool != "" {
		var err error
		place, err = pClient.GetQemuPlacableCluster(placeRequest)
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
		node, err := pClient.Cluster(qemu.Status.Cluster).GetQemuPlacableNode(placeRequest)
		if err != nil {
			return qemu, fmt.Errorf("cannot find available node: %s", err)
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
		qemu.Status.Status = v1alpha1.STATUS_QEMU_EMPTY
		return qemu, fmt.Errorf("cannot build qemu config: %s", err)
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Create(qemuConfig)
	if err != nil {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_EMPTY
		return qemu, fmt.Errorf("cannot create qemu: %s", err)
	}

	qemuConfig, err = createQemuDisks(pClient, qemu, qemuConfig)
	if err != nil {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
		return qemu, fmt.Errorf("cannot create qemu disk: %s", err)
	}

	pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().SetConfig(qemuConfig)
	if err != nil {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
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

	for ifaceName, iface := range qemu.Spec.Network {
		if iface.Mac == "" {
			if ifaceCurrentMacs[ifaceName] == "" {
				if iface.Tag == 0 {
					result[ifaceName] = fmt.Sprintf("%s,bridge=%s", iface.Model, iface.Bridge)
				} else {
					result[ifaceName] = fmt.Sprintf("%s,bridge=%s,tag=%d", iface.Model, iface.Bridge, iface.Tag)
				}
			} else {
				if iface.Tag == 0 {
					result[ifaceName] = fmt.Sprintf("%s=%s,bridge=%s", iface.Model, ifaceCurrentMacs[ifaceName], iface.Bridge)
				} else {
					result[ifaceName] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, ifaceCurrentMacs[ifaceName], iface.Bridge, iface.Tag)
				}
			}
		} else {
			if iface.Tag == 0 {
				result[ifaceName] = fmt.Sprintf("%s=%s,bridge=%s", iface.Model, iface.Mac, iface.Bridge)
			} else {
				result[ifaceName] = fmt.Sprintf("%s=%s,bridge=%s,tag=%d", iface.Model, iface.Mac, iface.Bridge, iface.Tag)
			}
		}
	}

	for k, v := range qemu.Spec.Options {
		result[k] = v
	}

	var tags []string
	tags = append(tags, qemu.Spec.Tags...)

	if qemu.Spec.AntiAffinity != "" {
		tags = append(tags, fmt.Sprintf("anti-affinity.%s", qemu.Spec.AntiAffinity))
	}

	if len(tags) > 0 {
		result["tags"] = strings.Join(tags, ";")
	}

	return result, nil
}

func createQemuDisks(pClient *proxmox.Client, qemu v1alpha1.Qemu, qemuConfig proxmox.QemuConfig) (proxmox.QemuConfig, error) {
	disksConfig, err := buildDisksConfig(qemu)
	if err != nil {
		return qemuConfig, fmt.Errorf("cannot build disks config: %s", err)
	}

	for _, diskConfig := range disksConfig {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).DiskCreate(diskConfig)
		if err != nil {
			return qemuConfig, fmt.Errorf("cannot create qemu disk: %s", err)
		}
		qemuConfig[diskConfig.Name] = fmt.Sprintf("%s:%s,size=%s", diskConfig.Storage, diskConfig.Filename, diskConfig.Size)
	}

	return qemuConfig, nil
}

func buildDisksConfig(qemu v1alpha1.Qemu) ([]proxmox.DiskConfig, error) {
	var disksConfig []proxmox.DiskConfig

	for diskName, disk := range qemu.Spec.Disk {
		diskConfig, err := buildDiskConfig(qemu, diskName, disk)
		if err != nil {
			return disksConfig, fmt.Errorf("cannot build disk config: %s %s", disk, err)
		}

		disksConfig = append(disksConfig, diskConfig)
	}
	return disksConfig, nil
}

func buildDiskConfig(qemu v1alpha1.Qemu, diskName string, disk v1alpha1.QemuDisk) (proxmox.DiskConfig, error) {
	var diskConfig proxmox.DiskConfig
	rDiskNum := regexp.MustCompile(`^[a-z]+(\d+)$`)
	diskNum := rDiskNum.FindStringSubmatch(diskName)
	if len(diskNum) != 2 {
		return diskConfig, fmt.Errorf("cannot extract disk num: %s", diskName)
	}

	filename := fmt.Sprintf("vm-%d-disk-%s", qemu.Status.VmId, diskNum[1])
	diskConfig = proxmox.DiskConfig{
		Name:     diskName,
		Node:     qemu.Status.Node,
		VmId:     qemu.Status.VmId,
		Filename: filename,
		Size:     disk.Size,
		Storage:  disk.Storage,
	}

	return diskConfig, nil
}

func buildPlaceRequest(qemu v1alpha1.Qemu) proxmox.PlaceRequest {
	var result proxmox.PlaceRequest

	result.Name = qemu.Metadata.Name
	result.CPU = qemu.Spec.CPU.Sockets + qemu.Spec.CPU.Cores
	result.Mem = qemu.Spec.Memory.Size
	result.AntiAffinity = qemu.Spec.AntiAffinity

	return result
}

func getQemuPowerStatus(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	qemuStatus, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetStatus(qemu.Status.VmId)
	if err != nil {
		qemu.Status.Power = v1alpha1.STATUS_POWER_UNKNOWN
		return qemu, fmt.Errorf("cannot get qemu power status: %s", err)
	}

	switch qemuStatus.Data.Status {
	case proxmox.STATUS_RUNNING:
		qemu.Status.Power = v1alpha1.STATUS_POWER_ON
	case proxmox.STATUS_STOPPED:
		qemu.Status.Power = v1alpha1.STATUS_POWER_OFF
	default:
		qemu.Status.Power = v1alpha1.STATUS_POWER_UNKNOWN
	}

	return qemu, nil
}

func getQemuNetStatus(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, fmt.Errorf("cannot get qemu config from proxmox: %s", err)
	}

	var ifacesConfig []v1alpha1.QemuStatusNetwork
	rExp := "(.{2}:.{2}:.{2}:.{2}:.{2}:.{2})"
	r := regexp.MustCompile(rExp)
	for ifaceName, _ := range qemu.Spec.Network {
		var ifaceConfig v1alpha1.QemuStatusNetwork
		ifaceConfig.Name = ifaceName
		macData := r.FindStringSubmatch(fmt.Sprintf("%s", currentConfig[ifaceName]))
		if len(macData) > 0 {
			ifaceConfig.Mac = macData[0]
		}

		ifacesConfig = append(ifacesConfig, ifaceConfig)
	}
	qemu.Status.Net = ifacesConfig

	return qemu, nil
}

func deleteQemu(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	if qemu.Status.Cluster == "" || qemu.Status.Node == "" || qemu.Status.VmId == 0 {
		return qemu, fmt.Errorf("unknown qemu status")
	}

	if qemu.Spec.Autostop {
		err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Stop(qemu.Status.VmId)
		if err != nil {
			return qemu, fmt.Errorf("cannot stop qemu: %s", err)
		}
	}

	qemu, err := getQemuPowerStatus(pClient, qemu)
	if err != nil {
		return qemu, fmt.Errorf("cannot get qemu power status: %s", err)
	}
	if qemu.Status.Power == v1alpha1.STATUS_POWER_ON {
		return qemu, errors.New("waiting for qemu stop")
	} else {
		err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Delete(qemu.Status.VmId)
		if err != nil {
			return qemu, fmt.Errorf("cannot delete qemu: %s", err)
		}
	}

	return qemu, nil
}

func checkQemuSyncStatus(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	//Check deleting state
	if qemu.Metadata.DeletionTimestamp != "" {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_DELETING

		return qemu, nil
	}

	//Check config state
	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, fmt.Errorf("cannot get qemu current config: %s", err)
	}

	designConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		return qemu, fmt.Errorf("cannot build qemu config: %s", err)
	}

	var outOfSync bool
	for k, v := range designConfig {
		v := strings.TrimRight(fmt.Sprint(v), "\n") // FUCK YOU PROXMOX!

		if k == "node" || k == "vmid" {
			continue
		}

		if k == "tags" {
			designTags := strings.Split(v, ";")
			currentTags := strings.Split(fmt.Sprint(currentConfig["tags"]), ";")

			for _, tag := range designTags {
				if !slices.Contains(currentTags, tag) {
					log.Warnf("Qemu %s is out of sync, tag %s is not found", qemu.Metadata.Name, tag)
					qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
					return qemu, nil
				}
			}

			continue
		}

		if fmt.Sprint(currentConfig[k]) != v {
			log.Warnf("Qemu %s is out of sync, %s: %s != %s", qemu.Metadata.Name, k, currentConfig[k], v)
			outOfSync = true
		}
	}

	if outOfSync {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC
		return qemu, nil
	}

	pendingConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetPendingConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, fmt.Errorf("cannot get pending config: %s", err)
	}

	//Check Disks config
	rDiskSize := regexp.MustCompile(`^.+size=(.+),?$`)
	for diskName, disk := range qemu.Spec.Disk {
		var isDiskFound bool
		var isDiskChanged bool
		for param, paramValue := range currentConfig {
			if diskName == param {
				isDiskFound = true
				currentSize := rDiskSize.FindStringSubmatch(fmt.Sprint(paramValue))
				if len(currentSize) != 2 {
					return qemu, fmt.Errorf("cannot parse disk params: %s %s", param, paramValue)
				}
				if disk.Size != currentSize[1] {
					isDiskChanged = true
				}
			}
		}
		if !isDiskFound || isDiskChanged {
			log.Warnf("Qemu %s disk is not found or changed: %s", qemu.Metadata.Name, disk)
			qemu.Status.Status = v1alpha1.STATUS_QEMU_OUT_OF_SYNC

			return qemu, nil
		}
	}

	//Check pending state
	var isPending bool
	for _, v := range pendingConfig {
		if v.Pending != nil && v.Value != v.Pending {
			log.Warnf("Qemu %s is in pending state, %s: %v != %v", qemu.Metadata.Name, v.Key, v.Value, v.Pending)
			isPending = true
		}
	}

	if isPending {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_PENDING
		return qemu, nil
	}

	qemu.Status.Status = v1alpha1.STATUS_QEMU_SYNCED

	return qemu, nil
}

func setQemuConfig(pClient *proxmox.Client, qemu v1alpha1.Qemu) (v1alpha1.Qemu, error) {
	qemuConfig, err := buildQemuConfig(pClient, qemu)
	if err != nil {
		return qemu, fmt.Errorf("cannot build qemu config: %s", err)
	}

	currentConfig, err := pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().GetConfig(qemu.Status.VmId)
	if err != nil {
		return qemu, fmt.Errorf("cannot get qemu current config: %s", err)
	}
	rDiskSize := regexp.MustCompile(`^.+size=(.+),?$`)
	for diskName, disk := range qemu.Spec.Disk {
		var isDiskFound bool
		for param, paramValue := range currentConfig {
			if diskName == param {
				isDiskFound = true
				currentSize := rDiskSize.FindStringSubmatch(fmt.Sprint(paramValue))
				if len(currentSize) != 2 {
					return qemu, fmt.Errorf("cannot parse disk params: %s %s", param, paramValue)
				}
				if disk.Size != currentSize[1] {
					err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().Resize(qemu.Status.VmId, diskName, disk.Size)
					if err != nil {
						return qemu, fmt.Errorf("cannot resize qemu disk: %s", err)
					}
				}
			}
		}
		if !isDiskFound {
			log.Warnf("Qemu %s disk is not found: %s", qemu.Metadata.Name, disk)
			diskConfig, err := buildDiskConfig(qemu, diskName, disk)
			if err != nil {
				return qemu, err
			}

			err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).DiskCreate(diskConfig)
			if err != nil {
				log.Errorf("cannot create qemu disk: %s", err)
			}
			qemuConfig[diskName] = fmt.Sprintf("%s:%s,size=%s", diskConfig.Storage, diskConfig.Filename, diskConfig.Size)

		}
	}

	err = pClient.Cluster(qemu.Status.Cluster).Node(qemu.Status.Node).Qemu().SetConfig(qemuConfig)
	if err != nil {
		return qemu, fmt.Errorf("cannot set qemu config: %s", err)
	}

	qemu, err = checkQemuSyncStatus(pClient, qemu)
	if err != nil {
		qemu.Status.Status = v1alpha1.STATUS_QEMU_UNKNOWN
	}

	return qemu, nil
}
