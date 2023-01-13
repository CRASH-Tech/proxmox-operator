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

func PostReq(clusterApiConfig ApiConfig, apiPath string, data interface{}) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", clusterApiConfig.ApiTokenId, clusterApiConfig.ApiTokenSecret)).
		SetBody(data).
		Post(fmt.Sprintf("%s/%s", clusterApiConfig.ApiUrl, apiPath))

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return err
}

func DeleteReq(clusterApiConfig ApiConfig, apiPath string) error {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", clusterApiConfig.ApiTokenId, clusterApiConfig.ApiTokenSecret)).
		Delete(fmt.Sprintf("%s/%s", clusterApiConfig.ApiUrl, apiPath))

	if resp.IsError() {
		return fmt.Errorf("proxmox api error: %d %s", resp.StatusCode(), resp.Body())
	}

	return err
}
