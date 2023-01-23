package proxmox

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	clusters map[string]ClusterApiConfig
}

func NewClient(clusters map[string]ClusterApiConfig) *Client {
	client := Client{
		clusters: clusters,
	}

	return &client
}

func (client *Client) Cluster(cluster string) *Cluster {
	apiConfig, isExists := client.clusters[cluster]
	if !isExists {
		log.Error("unknown cluster: %s", cluster)
	}

	restyClient := resty.New()
	restyClient.SetHeader("Content-Type", "application/json")
	restyClient.SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", apiConfig.ApiTokenId, apiConfig.ApiTokenSecret))

	result := Cluster{
		name:      cluster,
		apiCOnfig: apiConfig,
		resty:     restyClient,
	}

	return &result
}
