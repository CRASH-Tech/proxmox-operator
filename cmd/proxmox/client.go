package proxmox

import (
	"crypto/tls"
	"fmt"
	"sort"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	NODE_STATUS_ONLINE  = "online"
	NODE_STATUS_OFFLINE = "offline"
)

type Client struct {
	Clusters map[string]ClusterApiConfig
}

type QemuPlace struct {
	Cluster string
	Node    string
	VmId    int
	Found   bool
}

func NewClient(clusters map[string]ClusterApiConfig) *Client {
	client := Client{
		Clusters: clusters,
	}

	return &client
}

func (client *Client) Cluster(cluster string) *Cluster {
	apiConfig, isExists := client.Clusters[cluster]
	if !isExists {
		log.Error("unknown cluster: %s", cluster)
	}

	restyClient := resty.New()
	restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	restyClient.SetHeader("Content-Type", "application/json")
	restyClient.SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", apiConfig.ApiTokenId, apiConfig.ApiTokenSecret))

	result := Cluster{
		name:      cluster,
		apiCOnfig: apiConfig,
		resty:     restyClient,
	}

	return &result
}

func (client *Client) GetQemuPlacableCluster(cpu, mem int) (QemuPlace, error) {
	qemuCount := make(map[string]int)
	for cluster, _ := range client.Clusters {
		if count, err := client.Cluster(cluster).GetResourceCount(RESOURCE_QEMU); err == nil {
			qemuCount[cluster] = count
		}
	}

	keys := make([]string, 0, len(qemuCount))
	for k := range qemuCount {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return qemuCount[keys[i]] < qemuCount[keys[j]]
	})

	for _, cluster := range keys {
		if node, err := client.Cluster(cluster).GetQemuPlacableNode(cpu, mem); err == nil && node != "" {
			if vmId, err := client.Cluster(cluster).GetNextId(); err == nil {
				var result QemuPlace

				result.Cluster = cluster
				result.Node = node
				result.VmId = vmId

				return result, nil
			}
		}

	}

	return QemuPlace{}, fmt.Errorf("cannot find avialable cluster")
}

func (client *Client) GetQemuPlace(name string) (QemuPlace, error) {
	var place QemuPlace
	for cluster, _ := range client.Clusters {
		resources, err := client.Cluster(cluster).GetResources(RESOURCE_QEMU)
		if err != nil {
			return place, err
		}

		for _, resource := range resources {
			if resource.Name == name {
				place.Cluster = cluster
				place.Node = resource.Node
				place.VmId = resource.Vmid
				place.Found = true

				return place, nil
			}

		}
	}

	return place, nil
}
