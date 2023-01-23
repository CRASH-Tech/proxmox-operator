package proxmox

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type ApiConfig struct {
	ApiUrl         string `yaml:"api_url"`
	ApiTokenId     string `yaml:"api_token_id"`
	ApiTokenSecret string `yaml:"api_token_secret"`
}

type Client struct {
	clusters map[string]ApiConfig
}

func NewClient(clusters map[string]ApiConfig) *Client {
	client := Client{
		clusters: clusters,
	}

	return &client
}

// func (client *Client) getRestyClient() *resty.Client {
// 	restyClient := resty.New()
// 	restyClient.SetHeader("Content-Type", "application/json")
// 	restyClient.SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", client.Cluster(), clusterApiConfig.ApiTokenSecret))

// 	return restyClient
// }

func (client *Client) Cluster(cluster string) *Cluster {
	apiConfig, err := client.getApiConfig(cluster)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	result := Cluster{
		name:      cluster,
		apiConfig: apiConfig,
		resty:     getRestyClient(apiConfig.ApiTokenId, apiConfig.ApiTokenSecret),
	}

	return &result
}

func (client *Client) getApiConfig(clusterName string) (ApiConfig, error) {
	clusterApiConfig, isExists := client.clusters[clusterName]
	if !isExists {
		return clusterApiConfig, fmt.Errorf("unknown cluster: %s", clusterName)
	}

	return client.clusters[clusterName], nil
}

func getRestyClient(tokenId, tokenSecret string) *resty.Client {
	restyClient := resty.New()
	restyClient.SetHeader("Content-Type", "application/json")
	restyClient.SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", tokenId, tokenSecret))

	return restyClient
}
