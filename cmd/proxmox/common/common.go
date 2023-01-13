package common

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type ApiConfig struct {
	ApiUrl         string `yaml:"api_url"`
	ApiTokenId     string `yaml:"api_token_id"`
	ApiTokenSecret string `yaml:"api_token_secret"`
}

func getClient(clusterApiConfig ApiConfig) *resty.Client {
	client := resty.New()
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", clusterApiConfig.ApiTokenId, clusterApiConfig.ApiTokenSecret))

	return client
}

func PostReq(clusterApiConfig ApiConfig, apiPath string, data interface{}) error {
	client := getClient(clusterApiConfig)
	resp, err := client.R().
		SetBody(data).
		Post(fmt.Sprintf("%s/%s", clusterApiConfig.ApiUrl, apiPath))

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return err
}

func DeleteReq(clusterApiConfig ApiConfig, apiPath string) error {
	client := getClient(clusterApiConfig)
	resp, err := client.R().
		Delete(fmt.Sprintf("%s/%s", clusterApiConfig.ApiUrl, apiPath))

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return err
}
